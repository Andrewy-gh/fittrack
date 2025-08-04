import { useState } from 'react';
import { createFileRoute, Link } from '@tanstack/react-router';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Calendar,
  Dumbbell,
  Edit,
  Eye,
  Filter,
  MoreHorizontal,
  Plus,
  Search,
  Target,
  Trash2,
  TrendingUp,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import type { ExerciseOption } from '@/lib/types';
import { fetchExerciseOptions } from '@/lib/api/exercises';
import { Input } from '@/components/ui/input';

export const Route = createFileRoute('/_auth/exercises/')({
  loader: async ({ context }): Promise<ExerciseOption[]> => {
    const user = context.user;
    if (!user) {
      throw new Error('User not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
    const exercises = await fetchExerciseOptions(accessToken);
    return exercises;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const exercises = Route.useLoaderData();
  return <ExercisesDisplay exercises={exercises} />;
}

function ExercisesDisplay({ exercises }: { exercises: ExerciseOption[] }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedExercise, setSelectedExercise] =
    useState<ExerciseOption | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: '2-digit',
      year: 'numeric',
    });
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });
  };

  const getExerciseCategory = (name: string) => {
    const lowerName = name.toLowerCase();
    if (lowerName.includes('press') || lowerName.includes('bench'))
      return 'chest';
    if (lowerName.includes('row') || lowerName.includes('pull')) return 'back';
    if (lowerName.includes('squat') || lowerName.includes('lunge'))
      return 'legs';
    if (lowerName.includes('curl') || lowerName.includes('tricep'))
      return 'arms';
    if (lowerName.includes('deadlift')) return 'compound';
    if (lowerName.includes('lateral') || lowerName.includes('delt'))
      return 'shoulders';
    return 'general';
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case 'chest':
        return 'bg-red-500/20 text-red-500';
      case 'back':
        return 'bg-blue-500/20 text-blue-500';
      case 'legs':
        return 'bg-green-500/20 text-green-500';
      case 'arms':
        return 'bg-purple-500/20 text-purple-500';
      case 'shoulders':
        return 'bg-yellow-500/20 text-yellow-500';
      case 'compound':
        return 'bg-orange-500/20 text-orange-500';
      default:
        return 'bg-neutral-500/20 text-neutral-300';
    }
  };

  const filteredExercises = exercises.filter((exercise) =>
    exercise.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Calculate stats
  const totalExercises = exercises.length;
  const recentlyAdded = exercises.filter((ex) => {
    const exerciseDate = new Date(ex.created_at);
    const weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return exerciseDate >= weekAgo;
  }).length;

  const modifiedExercises = exercises.filter(
    (ex) => ex.updated_at !== null
  ).length;
  const categories = exercises.reduce(
    (acc, exercise) => {
      const category = getExerciseCategory(exercise.name);
      acc[category] = (acc[category] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>
  );

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-wider">
            EXERCISE DATABASE
          </h1>
          <p className="text-sm text-neutral-400">
            Movement library and exercise management
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            className="bg-orange-500 hover:bg-orange-600 text-white"
            onClick={() => setShowAddModal(true)}
          >
            <Plus className="w-4 h-4 mr-2" />
            New Exercise
          </Button>
          <Button className="bg-orange-500 hover:bg-orange-600 text-white">
            <Filter className="w-4 h-4 mr-2" />
            Filter
          </Button>
        </div>
      </div>

      {/* Search and Stats */}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
        <Card className="lg:col-span-2 bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-neutral-400" />
              <Input
                placeholder="Search exercises..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 bg-neutral-800 border-neutral-600 text-white placeholder-neutral-400"
              />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  TOTAL EXERCISES
                </p>
                <p className="text-2xl font-bold text-white font-mono">
                  {totalExercises}
                </p>
              </div>
              <Dumbbell className="w-8 h-8 text-white" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  RECENT ADDS
                </p>
                <p className="text-2xl font-bold text-orange-500 font-mono">
                  {recentlyAdded}
                </p>
              </div>
              <Calendar className="w-8 h-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  MODIFIED
                </p>
                <p className="text-2xl font-bold text-white font-mono">
                  {modifiedExercises}
                </p>
              </div>
              <TrendingUp className="w-8 h-8 text-white" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Exercise Categories Overview */}
      <Card className="bg-neutral-900 border-neutral-700">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-neutral-300 tracking-wider">
            MOVEMENT CATEGORIES
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-4">
            {Object.entries(categories).map(([category, count]) => (
              <div key={category} className="text-center">
                <div className="text-2xl font-bold text-white font-mono">
                  {count}
                </div>
                <div className="text-xs text-neutral-400 uppercase tracking-wider">
                  {category}
                </div>
                <Badge className={`${getCategoryColor(category)} mt-1`}>
                  {category.toUpperCase()}
                </Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Exercise List */}
      <Card className="bg-neutral-900 border-neutral-700">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-neutral-300 tracking-wider">
            EXERCISE REGISTRY
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-neutral-700">
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                    EXERCISE ID
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                    MOVEMENT NAME
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                    CATEGORY
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                    CREATED
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                    STATUS
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                    ACTIONS
                  </th>
                </tr>
              </thead>
              <tbody>
                {filteredExercises.map((exercise, index) => {
                  const category = getExerciseCategory(exercise.name);
                  return (
                    <tr
                      key={exercise.id}
                      className={`border-b border-neutral-800 hover:bg-neutral-800 transition-colors cursor-pointer ${
                        index % 2 === 0 ? 'bg-neutral-900' : 'bg-neutral-850'
                      }`}
                      onClick={() => setSelectedExercise(exercise)}
                    >
                      <td className="py-3 px-4 text-sm text-white font-mono">
                        EX-{exercise.id.toString().padStart(3, '0')}
                      </td>
                      <td className="py-3 px-4 text-sm text-white">
                        {exercise.name}
                      </td>
                      <td className="py-3 px-4">
                        <Badge className={getCategoryColor(category)}>
                          {category.toUpperCase()}
                        </Badge>
                      </td>
                      <td className="py-3 px-4">
                        <div className="text-sm text-neutral-300 font-mono">
                          {formatDate(exercise.created_at)}
                        </div>
                        <div className="text-xs text-neutral-500 font-mono">
                          {formatTime(exercise.created_at)}
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${exercise.updated_at ? 'bg-orange-500' : 'bg-white'}`}
                          ></div>
                          <span className="text-xs text-neutral-300 uppercase tracking-wider">
                            {exercise.updated_at ? 'MODIFIED' : 'ORIGINAL'}
                          </span>
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="text-neutral-400 hover:text-orange-500"
                        >
                          <MoreHorizontal className="w-4 h-4" />
                        </Button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Exercise Detail Modal */}
      {selectedExercise && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="bg-neutral-900 border-neutral-700 w-full max-w-2xl">
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-lg font-bold text-white tracking-wider">
                  {selectedExercise.name.toUpperCase()}
                </CardTitle>
                <p className="text-sm text-neutral-400 font-mono">
                  EX-{selectedExercise.id.toString().padStart(3, '0')}
                </p>
              </div>
              <Button
                variant="ghost"
                onClick={() => setSelectedExercise(null)}
                className="text-neutral-400 hover:text-white"
              >
                ✕
              </Button>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    CATEGORY
                  </p>
                  <Badge
                    className={getCategoryColor(
                      getExerciseCategory(selectedExercise.name)
                    )}
                  >
                    {getExerciseCategory(selectedExercise.name).toUpperCase()}
                  </Badge>
                </div>
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    STATUS
                  </p>
                  <div className="flex items-center gap-2">
                    <div
                      className={`w-2 h-2 rounded-full ${selectedExercise.updated_at ? 'bg-orange-500' : 'bg-white'}`}
                    ></div>
                    <span className="text-sm text-white uppercase tracking-wider">
                      {selectedExercise.updated_at ? 'MODIFIED' : 'ORIGINAL'}
                    </span>
                  </div>
                </div>
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    CREATED
                  </p>
                  <p className="text-sm text-white font-mono">
                    {formatDate(selectedExercise.created_at)}
                  </p>
                  <p className="text-xs text-neutral-400 font-mono">
                    {formatTime(selectedExercise.created_at)}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    LAST MODIFIED
                  </p>
                  <p className="text-sm text-white font-mono">
                    {selectedExercise.updated_at
                      ? formatDate(selectedExercise.updated_at)
                      : 'Never'}
                  </p>
                </div>
              </div>

              <div>
                <p className="text-xs text-neutral-400 tracking-wider mb-2">
                  EXERCISE DETAILS
                </p>
                <div className="bg-neutral-800 border border-neutral-700 rounded p-3">
                  <p className="text-sm text-white leading-relaxed">
                    Movement pattern: {selectedExercise.name}
                    <br />
                    Primary muscle group:{' '}
                    {getExerciseCategory(selectedExercise.name)}
                    <br />
                    Exercise classification: Resistance training
                  </p>
                </div>
              </div>

              <div className="flex gap-2 pt-4">
                <Button
                  className="bg-orange-500 hover:bg-orange-600 text-white"
                  asChild
                >
                  <Link
                    to="/exercises/$exerciseId"
                    params={{ exerciseId: selectedExercise.id }}
                  >
                    <Edit className="w-4 h-4 mr-2" />
                    Edit Exercise
                  </Link>
                </Button>
                <Button
                  variant="outline"
                  className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
                >
                  <Eye className="w-4 h-4 mr-2" />
                  View History
                </Button>
                <Button
                  variant="outline"
                  className="border-red-700 text-red-400 hover:bg-red-900/20 hover:text-red-300 bg-transparent"
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Add Exercise Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="bg-neutral-900 border-neutral-700 w-full max-w-md">
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-lg font-bold text-white tracking-wider">
                  ADD NEW EXERCISE
                </CardTitle>
                <p className="text-sm text-neutral-400">
                  Register new movement pattern
                </p>
              </div>
              <Button
                variant="ghost"
                onClick={() => setShowAddModal(false)}
                className="text-neutral-400 hover:text-white"
              >
                ✕
              </Button>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-xs text-neutral-400 tracking-wider mb-2 block">
                  EXERCISE NAME
                </label>
                <Input
                  placeholder="Enter exercise name..."
                  className="bg-neutral-800 border-neutral-600 text-white placeholder-neutral-400"
                />
              </div>

              <div>
                <label className="text-xs text-neutral-400 tracking-wider mb-2 block">
                  CATEGORY
                </label>
                <select className="w-full bg-neutral-800 border border-neutral-600 text-white rounded px-3 py-2">
                  <option value="">Select category...</option>
                  <option value="chest">Chest</option>
                  <option value="back">Back</option>
                  <option value="legs">Legs</option>
                  <option value="arms">Arms</option>
                  <option value="shoulders">Shoulders</option>
                  <option value="compound">Compound</option>
                  <option value="general">General</option>
                </select>
              </div>

              <div className="flex gap-2 pt-4">
                <Button className="bg-orange-500 hover:bg-orange-600 text-white flex-1">
                  <Target className="w-4 h-4 mr-2" />
                  Deploy Exercise
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setShowAddModal(false)}
                  className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
                >
                  Cancel
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
