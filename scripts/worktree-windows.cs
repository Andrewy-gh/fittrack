using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

// Anonymous, non-inheritable job: only this supervisor owns its lifetime.
public static class FitTrackJob {
    public static int Main(string[] args) {
        try {
            if (args[0] == "reserve") { Reserve(int.Parse(args[1])); return 0; }
            int code = Run(args[2], args.Skip(3).ToArray());
            File.WriteAllText(args[1], "drained");
            return code;
        } catch (Exception error) {
            // Do not echo command arguments or environment on failures.
            Console.Error.WriteLine("Windows supervisor failed: " + error.Message);
            return 125;
        }
    }
    [StructLayout(LayoutKind.Sequential)] struct Startup {
        public int cb; public IntPtr reserved, desktop, title;
        public int x, y, width, height, charsX, charsY, fill, flags;
        public short show, reservedSize; public IntPtr reservedBytes, input, output, error;
    }
    [StructLayout(LayoutKind.Sequential)] struct ProcessInfo { public IntPtr process, thread; public uint pid, tid; }
    [StructLayout(LayoutKind.Sequential)] struct StartupEx { public Startup startup; public IntPtr attributes; }
    [StructLayout(LayoutKind.Sequential)] struct BasicLimit {
        public long processTime, jobTime; public uint flags; public UIntPtr min, max;
        public uint active; public UIntPtr affinity; public uint priority, scheduling;
    }
    [StructLayout(LayoutKind.Sequential)] struct Limits {
        public BasicLimit basic;
        public ulong readOps, writeOps, otherOps, readBytes, writeBytes, otherBytes;
        public UIntPtr processMemory, jobMemory, peakProcess, peakJob;
    }
    [DllImport("kernel32.dll", SetLastError=true)] static extern IntPtr CreateJobObject(IntPtr attributes, string name);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool SetInformationJobObject(IntPtr job, int kind, ref Limits limits, int size);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool QueryInformationJobObject(IntPtr job, int kind, IntPtr info, int size, IntPtr returned);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool TerminateJobObject(IntPtr job, uint code);
    [DllImport("kernel32.dll", CharSet=CharSet.Unicode, SetLastError=true)] static extern bool CreateProcess(string app, StringBuilder command, IntPtr pa, IntPtr ta, bool inherit, uint flags, IntPtr env, string cwd, ref StartupEx startup, out ProcessInfo info);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool InitializeProcThreadAttributeList(IntPtr list, int count, int flags, ref IntPtr size);
    [DllImport("kernel32.dll", SetLastError=true)] static extern bool UpdateProcThreadAttribute(IntPtr list, uint flags, IntPtr attribute, IntPtr value, IntPtr size, IntPtr previous, IntPtr returned);
    [DllImport("kernel32.dll")] static extern void DeleteProcThreadAttributeList(IntPtr list);
    [DllImport("kernel32.dll")] static extern uint WaitForSingleObject(IntPtr handle, uint milliseconds);
    [DllImport("kernel32.dll")] static extern bool GetExitCodeProcess(IntPtr process, out uint code);
    [DllImport("kernel32.dll")] static extern bool CloseHandle(IntPtr handle);
    [DllImport("kernel32.dll")] static extern IntPtr GetStdHandle(int kind);
    static void Check(bool success) { if (!success) throw new Win32Exception(Marshal.GetLastWin32Error()); }

    // CommandLineToArgvW/CRT quoting, never a shell command string.
    static string Quote(string value) {
        var result = new StringBuilder("\""); int slashes = 0;
        foreach (char c in value) {
            if (c == '\\') { slashes++; continue; }
            result.Append('\\', c == '"' ? slashes * 2 + 1 : slashes);
            result.Append(c); slashes = 0;
        }
        return result.Append('\\', slashes * 2).Append('"').ToString();
    }
    public static int Run(string command, string[] args) {
        IntPtr job = CreateJobObject(IntPtr.Zero, null);
        if (job == IntPtr.Zero) throw new Win32Exception(Marshal.GetLastWin32Error());
        ProcessInfo child = new ProcessInfo();
        try {
            var limits = new Limits(); limits.basic.flags = 0x2000; // KILL_ON_JOB_CLOSE; no breakaway
            Check(SetInformationJobObject(job, 9, ref limits, Marshal.SizeOf(limits)));
            var startup = new StartupEx(); startup.startup.cb = Marshal.SizeOf(startup);
            startup.startup.flags = 0x100; // stdout/stderr inherited; stdin deliberately closed
            startup.startup.input = IntPtr.Zero; startup.startup.output = GetStdHandle(-11); startup.startup.error = GetStdHandle(-12);
            var line = new StringBuilder(Quote(command));
            foreach (string arg in args) line.Append(' ').Append(Quote(arg));
            // Windows 10+: atomic job assignment closes the create-suspended/assign crash gap.
            IntPtr size = IntPtr.Zero;
            InitializeProcThreadAttributeList(IntPtr.Zero, 1, 0, ref size);
            startup.attributes = Marshal.AllocHGlobal(size);
            IntPtr jobList = Marshal.AllocHGlobal(IntPtr.Size);
            bool initialized = false;
            try {
                Check(InitializeProcThreadAttributeList(startup.attributes, 1, 0, ref size)); initialized = true;
                Marshal.WriteIntPtr(jobList, job);
                Check(UpdateProcThreadAttribute(startup.attributes, 0, new IntPtr(0x0002000D), jobList, new IntPtr(IntPtr.Size), IntPtr.Zero, IntPtr.Zero));
                Check(CreateProcess(command, line, IntPtr.Zero, IntPtr.Zero, true, 0x00080000 | 0x08000000, IntPtr.Zero, null, ref startup, out child));
            } finally {
                if (initialized) DeleteProcThreadAttributeList(startup.attributes);
                Marshal.FreeHGlobal(startup.attributes); Marshal.FreeHGlobal(jobList);
            }
            // EOF also arrives when Node is forcibly killed: descendants cannot outlive it.
            Task<string> cancelled = Task.Run(() => Console.In.ReadLine());
            while (WaitForSingleObject(child.process, 50) == 258 && !cancelled.IsCompleted) {}
            uint code;
            Check(GetExitCodeProcess(child.process, out code));
            if (cancelled.IsCompleted) code = 130;
            Check(TerminateJobObject(job, 130)); // includes orphaned/background grandchildren
            IntPtr accounting = Marshal.AllocHGlobal(48);
            try {
                for (int i = 0; i < 100; i++) {
                    Check(QueryInformationJobObject(job, 1, accounting, 48, IntPtr.Zero));
                    if (Marshal.ReadInt32(accounting, 40) == 0) return unchecked((int)code);
                    Thread.Sleep(50);
                }
                throw new Exception("Owned job did not drain within five seconds.");
            } finally { Marshal.FreeHGlobal(accounting); }
        } finally {
            CloseHandle(job);
            if (child.thread != IntPtr.Zero) CloseHandle(child.thread);
            if (child.process != IntPtr.Zero) CloseHandle(child.process);
        }
    }
    public static void Reserve(int port) {
        var sockets = new List<Socket>();
        try {
            for (int p = port; p < port + 10; p++) {
                foreach (var address in new [] { IPAddress.Any, IPAddress.IPv6Any }) {
                    var socket = new Socket(address.AddressFamily, SocketType.Stream, ProtocolType.Tcp);
                    sockets.Add(socket);
                    socket.ExclusiveAddressUse = true; // Node exclusive:true only controls cluster sharing
                    if (address.AddressFamily == AddressFamily.InterNetworkV6) socket.DualMode = false;
                    try { socket.Bind(new IPEndPoint(address, p)); socket.Listen(1); }
                    catch (SocketException e) {
                        if (address.AddressFamily == AddressFamily.InterNetworkV6 && e.SocketErrorCode == SocketError.AddressFamilyNotSupported) continue;
                        Console.Out.WriteLine(e.SocketErrorCode == SocketError.AddressAlreadyInUse ? "EADDRINUSE" : "SOCKET_ERROR_" + e.ErrorCode);
                        return;
                    }
                }
            }
            Console.Out.WriteLine("READY"); Console.Out.Flush();
            Console.In.ReadLine();
        } finally { foreach (var socket in sockets) socket.Dispose(); }
    }
}
