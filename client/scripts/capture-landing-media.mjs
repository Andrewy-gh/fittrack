#!/usr/bin/env node
// Records the real app (guest demo mode) for the landing page "How it works"
// section and the README screenshots.
//
//   bun run build && bun run serve:test     # terminal 1
//   bun run media:capture                   # terminal 2 (needs ffmpeg on PATH)
//
// Output:
//   public/media/how-it-works/{log,repeat,review}-{light,dark}.{mp4,webp}
//   ../docs/media/app-screens-{light,dark}.png, ../docs/media/how-it-works.gif
import { chromium } from "playwright";
import { spawnSync } from "node:child_process";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const BASE_URL = process.env.CAPTURE_BASE_URL ?? "http://127.0.0.1:5173";
const CLIP_DIR = path.resolve(__dirname, "../public/media/how-it-works");
const README_DIR = path.resolve(__dirname, "../../docs/media");
const THEMES = ["light", "dark"];
const VIEWPORT = { width: 390, height: 844 };
const CLIP_WIDTH = 540; // 2x the phone frame's on-page width
const only = process.argv.slice(2);

// ---------- Seed: 12 weeks of plausible training in demo storage ----------

function buildSeed(now = new Date()) {
  const userId = "demo-user";
  const at = (daysAgo, hour = 18) => {
    const d = new Date(now);
    d.setDate(d.getDate() - daysAgo);
    d.setHours(hour, 15, 0, 0);
    return d;
  };
  const exercises = [
    "Barbell Squat",
    "Bench Press",
    "Deadlift",
    "Overhead Press",
    "Pull-ups",
    "Barbell Row",
  ].map((name, i) => ({
    id: i + 1,
    name,
    user_id: userId,
    created_at: at(86).toISOString(),
    updated_at: at(86).toISOString(),
  }));

  const offsets = [];
  // Alternating 3- and 2-day gaps ending yesterday: about 2.5 sessions a week.
  for (let d = 81, step = 3; d >= 1; d -= step, step = step === 3 ? 2 : 3) {
    offsets.push(d);
  }
  const workouts = [];
  const sets = [];
  const lerp = (from, to, t) => Math.round((from + (to - from) * t) / 5) * 5;
  const notes = {
    [offsets.length - 1]:
      "Bench: pause on chest, tuck elbows. Shoulder felt better. Try 195 next time.",
    [offsets.length - 2]:
      "Squat depth felt good. Brace harder on the last rep.",
    [offsets.length - 5]: "Deadlift grip slipping on set 3, use chalk.",
  };

  offsets.forEach((daysAgo, i) => {
    const workoutId = i + 1;
    const t = i / (offsets.length - 1);
    const upper = (offsets.length - 1 - i) % 2 === 0;
    const created = at(daysAgo).toISOString();
    workouts.push({
      id: workoutId,
      date: created.split("T")[0],
      notes: notes[i],
      workout_focus: upper ? "Upper Body" : "Lower Body",
      user_id: userId,
      created_at: created,
      updated_at: created,
    });
    const plan = upper
      ? [
          [2, lerp(165, 190, t), 5, 135],
          [4, lerp(95, 115, t), 6, 65],
          i % 4 < 2
            ? [6, lerp(135, 155, t), 8]
            : [5, undefined, 6 + Math.round(t * 4)],
        ]
      : [
          [1, lerp(205, 245, t), 5, 135],
          [3, lerp(245, 295, t), 5, 185],
        ];
    plan.forEach(([exerciseId, weight, reps, warmup], exerciseOrder) => {
      let setOrder = 0;
      const push = (w, r, type) =>
        sets.push({
          id: sets.length + 1,
          exercise_id: exerciseId,
          workout_id: workoutId,
          weight: w,
          reps: r,
          set_type: type,
          exercise_order: exerciseOrder,
          set_order: setOrder++,
          user_id: userId,
          created_at: created,
        });
      if (warmup) push(warmup, 8, "warmup");
      for (let s = 0; s < 3; s++) {
        push(weight, s === 2 && i % 3 === 1 ? reps - 1 : reps, "working");
      }
    });
  });

  return {
    "fittrack-demo-exercises": exercises,
    "fittrack-demo-workouts": workouts,
    "fittrack-demo-sets": sets,
  };
}

// ---------- Browser helpers ----------

// Shows a soft touch ripple wherever the script taps, so clips read as interaction.
function touchIndicatorScript() {
  addEventListener(
    "pointerdown",
    (e) => {
      const dot = document.createElement("div");
      dot.style.cssText = `position:fixed;left:${e.clientX - 22}px;top:${e.clientY - 22}px;width:44px;height:44px;border-radius:50%;background:rgba(127,127,127,.35);border:2px solid rgba(127,127,127,.8);pointer-events:none;z-index:2147483647;transition:transform .45s ease-out,opacity .45s ease-out`;
      document.documentElement.appendChild(dot);
      requestAnimationFrame(() => {
        dot.style.transform = "scale(1.5)";
        dot.style.opacity = "0";
      });
      setTimeout(() => dot.remove(), 600);
    },
    true,
  );
}

async function newAppPage(browser, theme, seed) {
  const context = await browser.newContext({
    viewport: VIEWPORT,
    deviceScaleFactor: 2,
    isMobile: true,
    hasTouch: true,
    colorScheme: theme,
    reducedMotion: "no-preference",
  });
  await context.addInitScript(
    ({ seed, theme }) => {
      if (sessionStorage.getItem("capture-seeded") !== "1") {
        localStorage.clear();
        for (const [key, value] of Object.entries(seed)) {
          localStorage.setItem(key, JSON.stringify(value));
        }
        sessionStorage.setItem("capture-seeded", "1");
      }
      localStorage.setItem("vite-ui-theme", theme);
    },
    { seed, theme },
  );
  await context.addInitScript(touchIndicatorScript);
  const page = await context.newPage();
  page.on("pageerror", (err) => console.warn(`  [page error] ${err.message}`));
  return page;
}

// Records CDP screencast frames (only emitted when pixels change) and keeps
// their timestamps so ffmpeg can rebuild real-time playback.
async function startScreencast(page) {
  const cdp = await page.context().newCDPSession(page);
  const frames = [];
  cdp.on("Page.screencastFrame", ({ data, metadata, sessionId }) => {
    frames.push({ data, t: metadata.timestamp });
    cdp.send("Page.screencastFrameAck", { sessionId }).catch(() => {});
  });
  await cdp.send("Page.startScreencast", {
    format: "jpeg",
    quality: 92,
    maxWidth: page.viewportSize().width * 2,
    maxHeight: page.viewportSize().height * 2,
  });
  return async () => {
    const end = Date.now() / 1000;
    await cdp.send("Page.stopScreencast");
    await cdp.detach();
    return { frames, end };
  };
}

function ffmpeg(args) {
  const res = spawnSync("ffmpeg", ["-y", "-loglevel", "error", ...args], {
    stdio: "inherit",
  });
  if (res.status !== 0) throw new Error(`ffmpeg failed: ${args.join(" ")}`);
}

async function encodeClip(
  { frames, end },
  outBase,
  posterAt,
  width = CLIP_WIDTH,
) {
  if (frames.length < 2)
    throw new Error(`only ${frames.length} frames captured`);
  const dir = await mkdtemp(path.join(tmpdir(), "fittrack-clip-"));
  const lines = [];
  for (let i = 0; i < frames.length; i++) {
    const file = path.join(dir, `f${String(i).padStart(5, "0")}.jpg`);
    await writeFile(file, Buffer.from(frames[i].data, "base64"));
    const next = i + 1 < frames.length ? frames[i + 1].t : end;
    lines.push(
      `file '${file.replaceAll("\\", "/")}'`,
      `duration ${Math.max(0.001, next - frames[i].t).toFixed(3)}`,
    );
  }
  lines.push(lines.at(-2)); // concat demuxer needs the last file repeated
  const list = path.join(dir, "list.txt");
  await writeFile(list, lines.join("\n"));

  ffmpeg([
    "-f",
    "concat",
    "-safe",
    "0",
    "-i",
    list,
    "-vf",
    `fps=30,scale=${width}:-2:flags=lanczos,format=yuv420p`,
    "-c:v",
    "libx264",
    "-crf",
    "26",
    "-preset",
    "slow",
    "-tune",
    "animation",
    "-movflags",
    "+faststart",
    "-an",
    `${outBase}.mp4`,
  ]);

  const duration = end - frames[0].t;
  const t = posterAt < 0 ? duration + posterAt : posterAt;
  ffmpeg([
    "-ss",
    t.toFixed(2),
    "-i",
    `${outBase}.mp4`,
    "-frames:v",
    "1",
    "-c:v",
    "libwebp",
    "-quality",
    "82",
    `${outBase}.webp`,
  ]);
  await rm(dir, { recursive: true, force: true });
  return duration;
}

const pause = (page, ms) => page.waitForTimeout(ms);

async function smoothScroll(page, by, ms) {
  await page.evaluate(
    ({ by, ms }) =>
      new Promise((resolve) => {
        const start = scrollY;
        const t0 = performance.now();
        const ease = (t) => (t < 0.5 ? 2 * t * t : 1 - (-2 * t + 2) ** 2 / 2);
        const step = (now) => {
          const p = Math.min(1, (now - t0) / ms);
          scrollTo(0, start + by * ease(p));
          if (p < 1) requestAnimationFrame(step);
          else resolve();
        };
        requestAnimationFrame(step);
      }),
    { by, ms },
  );
}

// ---------- The three flows ----------

const CLIPS = {
  // Step 1: log sets for an exercise.
  log: {
    url: "/workouts/new",
    ready: (page) =>
      page.getByRole("link", { name: /add exercise/i }).waitFor(),
    posterAt: -0.6,
    async run(page) {
      await pause(page, 700);
      await page.getByRole("link", { name: /add exercise/i }).click();
      await page.getByRole("heading", { name: /choose exercise/i }).waitFor();
      await pause(page, 900);
      await page
        .getByRole("button", { name: /bench press/i })
        .first()
        .click();
      await page.getByRole("button", { name: /add set/i }).waitFor();
      await pause(page, 1200);
      await page.getByRole("button", { name: /add set/i }).click();
      await page.getByRole("dialog").waitFor();
      await pause(page, 400);
      const [w, r] = [
        page.getByRole("spinbutton").first(),
        page.getByRole("spinbutton").nth(1),
      ];
      // Reps first: the dialog validates reps >= 1 as soon as any field changes.
      await r.fill("");
      await r.pressSequentially("5", { delay: 110 });
      await pause(page, 200);
      await w.fill("");
      await w.pressSequentially("195", { delay: 110 });
      await pause(page, 400);
      await page.getByRole("button", { name: /save set/i }).click();
      await page.getByRole("dialog").waitFor({ state: "detached" });
      await pause(page, 1100);
      for (let i = 0; i < 2; i++) {
        await page.getByRole("button", { name: /repeat last set/i }).click();
        await pause(page, 850);
      }
      await pause(page, 900);
    },
  },

  // Step 2: open the last session and repeat it as today's draft.
  repeat: {
    url: "/workouts",
    ready: (page) =>
      page
        .getByRole("link", { name: /upper body/i })
        .first()
        .waitFor(),
    // Open the workout once in-app before recording. On a cold query cache the
    // detail route suspends to the root and blanks the whole shell (app bug);
    // a user who viewed it earlier in the session sees it render instantly.
    async warm(page) {
      await page
        .getByRole("link", { name: /upper body/i })
        .first()
        .click();
      await page.getByRole("button", { name: /^repeat$/i }).waitFor();
      await page.goBack();
      await this.ready(page);
      await page.getByRole("link", { name: /new workout/i }).click();
      await page.getByRole("link", { name: /add exercise/i }).waitFor();
      await page.goBack();
      await this.ready(page);
    },
    posterAt: -0.6,
    async run(page) {
      await pause(page, 900);
      await page
        .getByRole("link", { name: /upper body/i })
        .first()
        .click();
      await page.getByRole("button", { name: /^repeat$/i }).waitFor();
      await pause(page, 1600);
      await page.getByRole("button", { name: /^repeat$/i }).click();
      await page.getByTestId("new-workout-exercise-card").first().waitFor();
      await pause(page, 1000);
      await page
        .getByRole("button", { name: /show last workout note/i })
        .click();
      await pause(page, 1400);
      await smoothScroll(page, 460, 1200);
      await pause(page, 1600);
    },
  },

  // Step 3: review consistency and volume over time.
  review: {
    url: "/analytics",
    ready: (page) => page.getByRole("heading", { name: /activity/i }).waitFor(),
    posterAt: 0.6, // before the first scroll: page title and stats
    async run(page) {
      await pause(page, 1200);
      await smoothScroll(page, 470, 1300);
      await pause(page, 1600);
      await smoothScroll(page, 400, 1300);
      await pause(page, 1800);
    },
  },
};

async function captureClips(browser, seed) {
  await mkdir(CLIP_DIR, { recursive: true });
  for (const theme of THEMES) {
    for (const [name, clip] of Object.entries(CLIPS)) {
      if (only.length && !only.includes(name)) continue;
      const page = await newAppPage(browser, theme, seed);
      await page.goto(BASE_URL + clip.url);
      await clip.ready(page);
      await clip.warm?.(page);
      await pause(page, 600); // let fonts and charts settle before recording
      const stop = await startScreencast(page);
      await clip.run(page);
      const recording = await stop();
      const duration = await encodeClip(
        recording,
        path.join(CLIP_DIR, `${name}-${theme}`),
        clip.posterAt,
      );
      console.log(
        `  ${name}-${theme}: ${recording.frames.length} frames, ${duration.toFixed(1)}s`,
      );
      await page.context().close();
    }
  }
}

// ---------- README artifacts ----------

async function captureReadmeScreens() {
  await mkdir(README_DIR, { recursive: true });
  for (const theme of THEMES) {
    const inputs = ["log", "repeat", "review"].map((n) =>
      path.join(CLIP_DIR, `${n}-${theme}.webp`),
    );
    const bg = theme === "dark" ? "0x0d1117" : "0xffffff";
    // Three posters side by side with gutters, on GitHub's page background.
    ffmpeg([
      ...inputs.flatMap((f) => ["-i", f]),
      "-filter_complex",
      `[0]pad=iw+48:ih+48:24:24:color=${bg}[a];[1]pad=iw+48:ih+48:24:24:color=${bg}[b];[2]pad=iw+48:ih+48:24:24:color=${bg}[c];[a][b][c]hstack=inputs=3,scale=1200:-2:flags=lanczos`,
      "-frames:v",
      "1",
      path.join(README_DIR, `app-screens-${theme}.png`),
    ]);
  }
  console.log("  docs/media/app-screens-{light,dark}.png");
}

// Records the landing page's scroll story on desktop and turns it into a GIF.
async function captureLandingGif(browser) {
  const context = await browser.newContext({
    viewport: { width: 1200, height: 760 },
    colorScheme: "dark",
  });
  await context.addInitScript(() =>
    localStorage.setItem("vite-ui-theme", "dark"),
  );
  const page = await context.newPage();
  await page.goto(BASE_URL + "/");
  const section = page.getByRole("region", { name: /log it\. repeat it\./i });
  await section.waitFor();
  // Page-y of each step's center, so the scroll parks each one mid-screen.
  const centers = await section
    .getByRole("heading", { level: 3 })
    .evaluateAll((headings) =>
      headings.map((heading) => {
        const rect = heading.closest("li").getBoundingClientRect();
        return rect.top + scrollY + rect.height / 2 - innerHeight / 2;
      }),
    );
  await page.evaluate((y) => scrollTo(0, y), centers[0]);
  await pause(page, 1500);
  const stop = await startScreencast(page);
  await pause(page, 4200);
  for (const y of centers.slice(1)) {
    const from = await page.evaluate(() => scrollY);
    await smoothScroll(page, y - from, 1300);
    await pause(page, 4200);
  }
  const recording = await stop();
  await context.close();

  const dir = await mkdtemp(path.join(tmpdir(), "fittrack-gif-"));
  const mp4 = path.join(dir, "landing");
  await encodeClip(recording, mp4, 0, 900);
  ffmpeg([
    "-i",
    `${mp4}.mp4`,
    "-vf",
    "crop=iw:ih-60:0:60,fps=10,scale=900:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128:stats_mode=diff[p];[s1][p]paletteuse=dither=bayer:bayer_scale=4:diff_mode=rectangle",
    path.join(README_DIR, "how-it-works.gif"),
  ]);
  await rm(dir, { recursive: true, force: true });
  console.log("  docs/media/how-it-works.gif");
}

const browser = await chromium.launch();
try {
  const res = await fetch(BASE_URL).catch(() => null);
  if (!res?.ok)
    throw new Error(
      `No app at ${BASE_URL}. Run "bun run build && bun run serve:test" first.`,
    );
  console.log("Recording clips");
  await captureClips(browser, buildSeed());
  if (!only.length || only.includes("readme")) {
    console.log("README artifacts");
    await captureReadmeScreens();
    await captureLandingGif(browser);
  }
} finally {
  await browser.close();
}
