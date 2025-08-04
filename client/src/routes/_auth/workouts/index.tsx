import { createFileRoute, Link } from '@tanstack/react-router';
// import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
// import { Button } from '@/components/ui/button';
// import { Input } from '@/components/ui/input';
// import { Badge } from '@/components/ui/badge';
import { fetchWorkouts } from '@/lib/api/workouts';
// import {
//   Search,
//   Filter,
//   Calendar,
//   Clock,
//   Target,
//   TrendingUp,
//   Plus,
//   MoreHorizontal,
// } from 'lucide-react';
// import {
//   Table,
//   TableBody,
//   TableCell,
//   TableHead,
//   TableHeader,
//   TableRow,
// } from '@/components/ui/table';
import { type WorkoutData } from '@/lib/api/workouts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Plus, Calendar, ChevronRight, Dumbbell } from 'lucide-react';

const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  if (date.toDateString() === today.toDateString()) {
    return 'Today';
  } else if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday';
  } else {
    return date.toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
    });
  }
};

const formatTime = (dateString: string) => {
  return new Date(dateString).toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
};

export function WorkoutsDisplay({ workouts }: { workouts: WorkoutData[] }) {
  const totalWorkouts = workouts.length;
  const thisWeekWorkouts = workouts.filter((workout) => {
    const workoutDate = new Date(workout.date);
    const weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return workoutDate >= weekAgo;
  }).length;

  const workoutTypes = workouts.reduce(
    (acc, workout) => {
      if (workout.notes) {
        acc[workout.notes] = (acc[workout.notes] || 0) + 1;
      }
      return acc;
    },
    {} as Record<string, number>
  );

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        {/* Header */}
        <div className="flex items-center justify-between pt-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Workouts</h1>
          </div>
          <Button size="sm" asChild>
            <Link to="/workouts/new-2">
              <Plus className="w-4 h-4 mr-2" />
              New Workout
            </Link>
          </Button>
        </div>
        {/* MARK: Summary Cards */}
        <div className="grid grid-cols-2 gap-4">
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Dumbbell className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Total Workouts</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {totalWorkouts}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Calendar className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">This Week</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {thisWeekWorkouts}
            </div>
          </Card>
        </div>
        {/* MARK: List */}
        <Card className="border-0 shadow-sm backdrop-blur-sm">
          <CardHeader>
            {/* <div className="flex items-center justify-between"> */}
            <CardTitle className="text-xl font-semibold">
              Recent Workouts
            </CardTitle>
            {/* <Button variant="ghost" size="sm">
                Show All
                <ChevronRight className="w-4 h-4 ml-1" />
              </Button>
            </div> */}
          </CardHeader>
          <CardContent className="space-y-3">
            {workouts.map((workout) => (
              <Link
                key={workout.id}
                to="/workouts/$workoutId"
                params={{
                  workoutId: workout.id,
                }}
                className="flex items-center justify-between py-4 rounded-xl cursor-pointer"
              >
                <div className="flex items-center space-x-4">
                  {workout.notes && (
                    <div className="flex items-center space-x-2">
                      <Badge
                        variant="outline"
                        className="border-border bg-muted text-xs"
                      >
                        {workout.notes.toUpperCase()}
                      </Badge>
                    </div>
                  )}
                  <div className="flex items-center space-x-2 text-sm mt-1">
                    <span className="uppercase text-muted-foreground">
                      {formatDate(workout.date)}
                    </span>
                    <span>•</span>
                    <span className="text-muted-foreground">
                      {formatTime(workout.created_at)}
                    </span>
                  </div>
                </div>
                <ChevronRight className="w-5 h-5 text-muted-foreground" />
              </Link>
            ))}
          </CardContent>
        </Card>

        {/* MARK: Distribution */}
        <Card className="p-4">
          <CardTitle className="text-xl font-semibold">
            Workout Distribution
          </CardTitle>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {Object.entries(workoutTypes).map(([type, count]) => (
                <div key={type} className="text-center p-4 rounded-xl">
                  <p className="font-semibold text-lg">{count}</p>
                  <p className="text-sm uppercase">{type}</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}

// function WorkoutsDisplay({ workouts }: { workouts: WorkoutData[] }) {
//   const [searchTerm, setSearchTerm] = useState('');
//   const [selectedWorkout, setSelectedWorkout] = useState<WorkoutData | null>(
//     null
//   );

//   const formatDate = (dateString: string) => {
//     return new Date(dateString).toLocaleDateString('en-US', {
//       month: 'short',
//       day: '2-digit',
//       year: 'numeric',
//     });
//   };

//   const formatTime = (dateString: string) => {
//     return new Date(dateString).toLocaleTimeString('en-US', {
//       hour: '2-digit',
//       minute: '2-digit',
//       hour12: false,
//     });
//   };

//   const getWorkoutType = (notes: string) => {
//     const lowerNotes = notes.toLowerCase();
//     if (
//       lowerNotes.includes('cardio') ||
//       lowerNotes.includes('run') ||
//       lowerNotes.includes('hiit')
//     )
//       return 'cardio';
//     if (
//       lowerNotes.includes('strength') ||
//       lowerNotes.includes('power') ||
//       lowerNotes.includes('lift')
//     )
//       return 'strength';
//     if (
//       lowerNotes.includes('yoga') ||
//       lowerNotes.includes('recovery') ||
//       lowerNotes.includes('flexibility')
//     )
//       return 'recovery';
//     if (lowerNotes.includes('circuit') || lowerNotes.includes('full body'))
//       return 'circuit';
//     return 'general';
//   };

//   const getTypeColor = (type: string) => {
//     switch (type) {
//       case 'cardio':
//         return 'bg-red-500/20 text-red-500';
//       case 'strength':
//         return 'bg-orange-500/20 text-orange-500';
//       case 'recovery':
//         return 'bg-white/20 text-white';
//       case 'circuit':
//         return 'bg-neutral-500/20 text-neutral-300';
//       default:
//         return 'bg-neutral-600/20 text-neutral-400';
//     }
//   };

//   const filteredWorkouts = workouts.filter(
//     (workout) =>
//       workout.notes?.toLowerCase().includes(searchTerm.toLowerCase()) ||
//       workout.id.toString().includes(searchTerm)
//   );

//   // Calculate stats
//   const totalWorkouts = workouts.length;
//   const thisWeekWorkouts = workouts.filter((w) => {
//     const workoutDate = new Date(w.date);
//     const weekAgo = new Date();
//     weekAgo.setDate(weekAgo.getDate() - 7);
//     return workoutDate >= weekAgo;
//   }).length;

//   const updatedWorkouts = workouts.filter((w) => w.updated_at !== null).length;
//   const workoutTypes = workouts.reduce(
//     (acc, workout) => {
//       const type = getWorkoutType(workout.notes ?? '');
//       acc[type] = (acc[type] || 0) + 1;
//       return acc;
//     },
//     {} as Record<string, number>
//   );

//   return (
//     <div className="p-6 space-y-6">
//       {/* Header */}
//       <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
//         <div>
//           <h1 className="text-2xl font-bold tracking-wider">
//             WORKOUT COMMAND
//           </h1>
//           <p className="text-sm">
//             Training session monitoring and analysis
//           </p>
//         </div>
//         <div className="flex gap-2">
//           <Button>
//             <Plus className="w-4 h-4 mr-2" />
//             New Session
//           </Button>
//           <Button>
//             <Filter className="w-4 h-4 mr-2" />
//             Filter
//           </Button>
//         </div>
//       </div>

//       {/* Search and Stats */}
//       <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
//         <Card className="lg:col-span-2">
//           <CardContent className="p-4">
//             <div className="relative">
//               <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4" />
//               <Input
//                 placeholder="Search workouts..."
//                 value={searchTerm}
//                 onChange={(e) => setSearchTerm(e.target.value)}
//                 className="pl-10"
//               />
//             </div>
//           </CardContent>
//         </Card>

//         <Card>
//           <CardContent className="p-4">
//             <div className="flex items-center justify-between">
//               <div>
//                 <p className="text-xs tracking-wider">
//                   TOTAL SESSIONS
//                 </p>
//                 <p className="text-2xl font-bold font-mono">
//                   {totalWorkouts}
//                 </p>
//               </div>
//               <Target className="w-8 h-8" />
//             </div>
//           </CardContent>
//         </Card>

//         <Card>
//           <CardContent className="p-4">
//             <div className="flex items-center justify-between">
//               <div>
//                 <p className="text-xs tracking-wider">
//                   THIS WEEK
//                 </p>
//                 <p className="text-2xl font-bold font-mono">
//                   {thisWeekWorkouts}
//                 </p>
//               </div>
//               <Calendar className="w-8 h-8" />
//             </div>
//           </CardContent>
//         </Card>

//         <Card>
//           <CardContent className="p-4">
//             <div className="flex items-center justify-between">
//               <div>
//                 <p className="text-xs tracking-wider">
//                   MODIFIED
//                 </p>
//                 <p className="text-2xl font-bold font-mono">
//                   {updatedWorkouts}
//                 </p>
//               </div>
//               <TrendingUp className="w-8 h-8" />
//             </div>
//           </CardContent>
//         </Card>
//       </div>

//       {/* Workout Types Overview */}
//       <Card>
//         <CardHeader>
//           <CardTitle className="text-sm font-medium tracking-wider">
//             TRAINING DISTRIBUTION
//           </CardTitle>
//         </CardHeader>
//         <CardContent>
//           <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-4">
//                 <Badge className="mt-1">
//                   {type.toUpperCase()}
//                 </Badge>
//               </div>
//             ))}
//           </div>
//         </CardContent>
//       </Card>

//       {/* MARK: Table */}
//       <Card>
//         <CardHeader>
//           <CardTitle className="text-sm font-medium tracking-wider">
//             TRAINING SESSIONS
//           </CardTitle>
//         </CardHeader>
//         <CardContent>
//           <div className="rounded-md border">
//             <Table>
//               <TableHeader>
//                 <TableRow>
//                   <TableHead className="hidden sm:table-cell">ID</TableHead>
//                   <TableHead className="font-medium tracking-wider">
//                     Date
//                   </TableHead>
//                   <TableHead className="hidden sm:table-cell">Time</TableHead>
//                   <TableHead className="hidden sm:table-cell">Type</TableHead>
//                   <TableHead>Notes</TableHead>
//                   <TableHead className="hidden sm:table-cell">Status</TableHead>
//                   <TableHead className="w-[50px]"></TableHead>
//                 </TableRow>
//               </TableHeader>
//               <TableBody>
//                 {filteredWorkouts.map((workout) => {
//                   const workoutType = getWorkoutType(workout.notes ?? '');
//                   return (
//                     <TableRow
//                       key={workout.id}
//                       onClick={() => setSelectedWorkout(workout)}
//                     >
//                       <TableCell className="hidden font-medium sm:table-cell">
//                         WO-{workout.id.toString().padStart(3, '0')}
//                       </TableCell>
//                       <TableCell>
//                         <div className="font-medium">
//                           {formatDate(workout.date)}
//                         </div>
//                         <div className="text-sm sm:hidden">
//                           {formatTime(workout.date)}
//                         </div>
//                       </TableCell>
//                       <TableCell className="hidden sm:table-cell">
//                         <div className="flex items-center gap-2">
//                           <Clock className="w-3 h-3" />
//                           <span>{formatTime(workout.date)}</span>
//                         </div>
//                       </TableCell>
//                       <TableCell className="hidden sm:table-cell">
//                         <Badge variant="outline">
//                           {workoutType.toUpperCase()}
//                         </Badge>
//                       </TableCell>
//                       <TableCell className="max-w-xs truncate">
//                         {workout.notes && (
//                           <Badge variant="outline">{workout.notes}</Badge>
//                         )}
//                       </TableCell>
//                       <TableCell className="hidden sm:table-cell">
//                         <div className="flex items-center gap-2">
//                           <div
//                             className={`w-2 h-2 rounded-full ${
//                               workout.updated_at ? '' : ''
//                             }`}
//                           />
//                           <span className="text-xs uppercase">
//                             {workout.updated_at ? 'MODIFIED' : 'ORIGINAL'}
//                           </span>
//                         </div>
//                       </TableCell>
//                       <TableCell>
//                         <Button
//                           variant="ghost"
//                           size="sm"
//                           onClick={(e) => {
//                             e.stopPropagation();
//                             setSelectedWorkout(workout);
//                           }}
//                         >
//                           <MoreHorizontal className="h-4 w-4" />
//                         </Button>
//                       </TableCell>
//                     </TableRow>
//                   );
//                 })}
//               </TableBody>
//             </Table>
//           </div>
//         </CardContent>
//       </Card>

//       {/* MARK: Modal */}
//       {selectedWorkout && (
//         <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
//           <Card className="w-full max-w-2xl">
//             <CardHeader className="flex flex-row items-center justify-between">
//               <div>
//                 <CardTitle className="text-lg font-bold tracking-wider">
//                   TRAINING SESSION WO-
//                   {selectedWorkout.id.toString().padStart(3, '0')}
//                 </CardTitle>
//                 <p className="text-sm font-mono">
//                   {formatDate(selectedWorkout.date)} at{' '}
//                   {formatTime(selectedWorkout.date)}
//                 </p>
//               </div>
//               <Button
//                 variant="ghost"
//                 onClick={() => setSelectedWorkout(null)}
//               >
//                 ✕
//               </Button>
//             </CardHeader>
//             <CardContent className="space-y-4">
//               <div className="grid grid-cols-2 gap-4">
//                 <div>
//                   <p className="text-xs tracking-wider mb-1">
//                     WORKOUT TYPE
//                   </p>
//                   <Badge>
//                     {getWorkoutType(selectedWorkout.notes ?? '').toUpperCase()}
//                   </Badge>
//                 </div>
//                 <div>
//                   <p className="text-xs tracking-wider mb-1">
//                     STATUS
//                   </p>
//                   <div className="flex items-center gap-2">
//                     <div
//                       className={`w-2 h-2 rounded-full ${
//                         selectedWorkout.updated_at ? '' : ''
//                       }`}
//                     ></div>
//                     <span className="text-sm uppercase tracking-wider">
//                       {selectedWorkout.updated_at ? 'MODIFIED' : 'ORIGINAL'}
//                     </span>
//                   </div>
//                 </div>
//                 <div>
//                   <p className="text-xs tracking-wider mb-1">
//                     CREATED
//                   </p>
//                   <p className="text-sm font-mono">
//                     {formatDate(selectedWorkout.created_at)}
//                   </p>
//                 </div>
//                 <div>
//                   <p className="text-xs tracking-wider mb-1">
//                     LAST MODIFIED
//                   </p>
//                   <p className="text-sm font-mono">
//                     {selectedWorkout.updated_at
//                       ? formatDate(selectedWorkout.updated_at)
//                       : 'Never'}
//                   </p>
//                 </div>
//               </div>

//               <div>
//                 <p className="text-xs tracking-wider mb-2">
//                   TRAINING NOTES
//                 </p>
//                 <div className="bg-neutral-800 border border-neutral-700 rounded p-3">
//                   <p className="text-sm leading-relaxed">
//                     {selectedWorkout.notes}
//                   </p>
//                 </div>
//               </div>

//               {/* <div className="flex gap-2 pt-4"> */}
//               <Button
//                 className=""
//                 asChild
//               >
//                 <Link
//                   to="/workouts/$workoutId"
//                   params={{
//                     workoutId: selectedWorkout.id,
//                   }}
//                 >
//                   View Session
//                 </Link>
//               </Button>
//               {/* <Button
//                   variant="outline"
//                   className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
//                 >
//                   Duplicate
//                 </Button>
//                 <Button
//                   variant="outline"
//                   className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
//                 >
//                   Export Data
//                 </Button> */}
//               {/* </div> */}
//             </CardContent>
//           </Card>
//         </div>
//       )}
//     </div>
//   );
// }

export const Route = createFileRoute('/_auth/workouts/')({
  loader: async ({ context }): Promise<WorkoutData[]> => {
    const user = context.user;
    if (!user) {
      throw new Error('User not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
    return fetchWorkouts(accessToken);
  },
  component: RouteComponent,
});

function RouteComponent() {
  const workouts = Route.useLoaderData();
  return <WorkoutsDisplay workouts={workouts} />;
}
