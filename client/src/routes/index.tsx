import { createFileRoute, Link } from '@tanstack/react-router';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  ArrowRight,
  BarChart3,
  Calendar,
  CheckCircle,
  Clock3,
  Copy,
  Dumbbell,
  Eye,
  NotebookText,
  Target,
  Zap,
} from 'lucide-react';
import { type CurrentInternalUser, type CurrentUser } from '@stackframe/react';
import { CustomUserButton } from '@/components/custom-user-button';
import { GuestUserButton } from '@/components/guest-user-button';

function HomePage({
  user,
}: {
  user: CurrentUser | CurrentInternalUser | null;
}) {
  const features = [
    {
      icon: Zap,
      title: 'Fast workout logging',
      description:
        'Add exercises, sets, reps, weight, and workout notes without fighting the form.',
      highlight: 'QUICK ENTRY',
    },
    {
      icon: Copy,
      title: 'Repeat what you already did',
      description:
        'Start from your last workout or reuse an existing session structure when that is faster than starting blank.',
      highlight: 'REPEAT LAST',
    },
    {
      icon: Eye,
      title: 'Progress you can see',
      description:
        'Review exercise history, recent sets, volume, charts, and past notes in the same place.',
      highlight: 'CLEAR HISTORY',
    },
    {
      icon: Calendar,
      title: 'Consistency without pressure',
      description:
        'See workouts this week, active days this month, and weekly averages without guilt-heavy streak mechanics.',
      highlight: 'PLAIN SUMMARY',
    },
  ];

  const groundedProof = [
    'Track exercises, sets, reps, weight, and workout notes.',
    'Set simple per-exercise targets for weight, reps, or weekly frequency.',
    'Bring the last workout note back into view when you start logging again.',
    'Review workout history and exercise progress without digging through old sessions.',
  ];

  return (
    <div className="min-h-screen">
      <nav className="border-b border-border bg-background/90 backdrop-blur-sm">
        <div className="flex items-center justify-between px-2 py-4">
          <div className="flex items-center gap-2">
            <Dumbbell className="h-6 w-6 text-primary" />
            <span className="text-xl font-bold tracking-wide text-foreground">
              FITTRACK
            </span>
          </div>
          {user ? <CustomUserButton /> : <GuestUserButton />}
        </div>
      </nav>

      <section className="px-6 pb-16 pt-24">
        <div className="mx-auto max-w-7xl">
          <div className="mb-12 space-y-6 text-center">
            <Badge className="mb-6 bg-primary/15 px-4 py-2 text-primary">
              <Clock3 className="mr-2 h-4 w-4" />
              Built for fast logging and useful history
            </Badge>
            <h1 className="text-5xl font-bold leading-snug tracking-wide text-foreground md:text-7xl">
              Log workouts fast.
              <br />
              See progress clearly.
            </h1>
            <p className="mx-auto max-w-3xl text-lg leading-relaxed text-muted-foreground">
              Track exercises, sets, reps, weight, and workout notes, then
              review your history, exercise progress, and consistency summaries
              in one place.
            </p>
            <div className="mb-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
              <Button
                size="lg"
                className="bg-primary px-8 py-4 text-primary-foreground hover:bg-primary/90"
                asChild
              >
                <Link to="/workouts" preload={false}>
                  <Dumbbell className="mr-1 h-5 w-5" />
                  {user ? 'Open Workouts' : 'Try the App'}
                </Link>
              </Button>
              <Button
                size="lg"
                variant="outline"
                className="border-border bg-transparent px-8 py-4 text-muted-foreground hover:bg-background/50"
                asChild
              >
                <Link to="/exercises" preload={false}>
                  Browse Exercise History
                </Link>
              </Button>
            </div>
          </div>

          <div className="grid grid-cols-1 items-center gap-12 lg:grid-cols-2">
            <div className="mx-auto w-full max-w-md space-y-8">
              <div className="space-y-4">
                <h2 className="text-xl font-medium text-muted-foreground">
                  A workout log that stays useful after you hit save
                </h2>
                <p className="leading-relaxed text-muted-foreground">
                  FitTrack is built around repeatable training: log the session,
                  reuse the structure next time, check the last cue, and keep
                  your history readable.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <Card className="p-4">
                  <CardContent className="space-y-3 p-0">
                    <div className="text-xs uppercase tracking-wide text-muted-foreground">
                      Track
                    </div>
                    <div className="text-2xl font-bold text-primary">
                      Sets, reps, weight
                    </div>
                    <div className="text-sm text-muted-foreground">
                      Plus workout focus and notes.
                    </div>
                  </CardContent>
                </Card>
                <Card className="p-4">
                  <CardContent className="space-y-3 p-0">
                    <div className="text-xs uppercase tracking-wide text-muted-foreground">
                      Review
                    </div>
                    <div className="text-2xl font-bold text-foreground">
                      History, charts, notes
                    </div>
                    <div className="text-sm text-muted-foreground">
                      Exercise detail and workout history stay connected.
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>

            <div className="flex justify-center">
              <div className="relative w-full max-w-sm overflow-hidden rounded-[2rem] border border-border bg-card p-6 shadow-2xl">
                <div className="mb-6 text-center">
                  <div className="mb-2 text-xs text-muted-foreground">
                    TODAY&apos;S WORKOUT
                  </div>
                  <div className="text-2xl font-bold text-foreground">
                    Upper Body
                  </div>
                </div>

                <div className="space-y-3">
                  <div className="rounded-xl border border-border bg-background/60 p-3">
                    <div className="mb-2 flex items-center justify-between">
                      <span className="text-sm font-medium text-foreground">
                        Repeat Last Workout
                      </span>
                      <Badge className="bg-primary/15 text-primary">
                        1 tap
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Load the last structure instead of rebuilding it.
                    </p>
                  </div>

                  <div className="rounded-xl border border-border bg-background/60 p-3">
                    <div className="mb-2 flex items-center justify-between">
                      <span className="text-sm font-medium text-foreground">
                        Bench Press
                      </span>
                      <Badge variant="outline">Goal 205 lb • 5 reps</Badge>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Recent sets and exercise history stay one tap away.
                    </p>
                  </div>

                  <div className="rounded-xl border border-border bg-background/60 p-3">
                    <div className="mb-2 flex items-center justify-between">
                      <span className="text-sm font-medium text-foreground">
                        This Week
                      </span>
                      <span className="text-sm font-semibold text-primary">
                        3 workouts
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      8 active days this month • 2.5 avg / week
                    </p>
                  </div>

                  <Button
                    className="mt-4 w-full bg-primary px-8 py-4 text-primary-foreground hover:bg-primary/90"
                    asChild
                  >
                    <Link to="/workouts" preload={false}>
                      <ArrowRight className="mr-1 h-5 w-5" />
                      Start Logging
                    </Link>
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="bg-background px-6 py-16">
        <div className="mx-auto max-w-7xl">
          <div className="mb-16 text-center">
            <h2 className="mb-4 text-4xl font-bold leading-snug tracking-wide text-foreground md:text-6xl">
              Grounded features,
              <br />
              <span className="text-muted-foreground">not product theater.</span>
            </h2>
            <p className="text-xl text-muted-foreground">
              FitTrack stays focused on the parts of training people actually
              return to.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-8 md:grid-cols-2">
            {features.map((feature) => (
              <Card key={feature.title} className="border border-border p-6">
                <CardContent className="p-0">
                  <div className="flex items-start gap-4">
                    <div className="rounded bg-primary/20 p-3">
                      <feature.icon className="h-6 w-6 text-primary" />
                    </div>
                    <div className="flex-1">
                      <div className="mb-2 flex items-center gap-2">
                        <h3 className="text-lg font-bold text-foreground">
                          {feature.title}
                        </h3>
                        <Badge className="bg-primary/20 text-xs text-primary">
                          {feature.highlight}
                        </Badge>
                      </div>
                      <p className="text-muted-foreground">
                        {feature.description}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      <section className="px-6 py-16">
        <div className="mx-auto grid max-w-7xl grid-cols-1 items-center gap-16 lg:grid-cols-2">
          <div>
            <h2 className="mb-6 text-4xl font-bold leading-snug tracking-wide text-foreground md:text-6xl">
              Useful history for
              <br />
              <span className="text-muted-foreground">real training.</span>
            </h2>
            <p className="mb-8 text-xl leading-relaxed text-muted-foreground">
              The app keeps enough context around each session that your next
              workout can start from something concrete.
            </p>

            <div className="space-y-4 text-muted-foreground">
              {groundedProof.map((item) => (
                <div key={item} className="flex items-center gap-3">
                  <CheckCircle className="h-5 w-5 text-primary" />
                  <span>{item}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="grid gap-4">
            <Card className="border border-border p-6">
              <CardContent className="p-0">
                <div className="mb-4 flex items-center gap-3">
                  <NotebookText className="h-5 w-5 text-primary" />
                  <h3 className="text-lg font-semibold">Workout notes resurface</h3>
                </div>
                <p className="text-sm leading-6 text-muted-foreground">
                  Last workout notes come back into view when you start a new
                  session, so important context is not trapped in old entries.
                </p>
              </CardContent>
            </Card>

            <Card className="border border-border p-6">
              <CardContent className="p-0">
                <div className="mb-4 flex items-center gap-3">
                  <Target className="h-5 w-5 text-primary" />
                  <h3 className="text-lg font-semibold">Simple exercise goals</h3>
                </div>
                <p className="text-sm leading-6 text-muted-foreground">
                  Set a target weight, reps, or weekly frequency for an
                  exercise. No coaching layer, no recommendations, just a clear
                  target beside the lift.
                </p>
              </CardContent>
            </Card>

            <Card className="border border-border p-6">
              <CardContent className="p-0">
                <div className="mb-4 flex items-center gap-3">
                  <BarChart3 className="h-5 w-5 text-primary" />
                  <h3 className="text-lg font-semibold">Consistency summaries</h3>
                </div>
                <p className="text-sm leading-6 text-muted-foreground">
                  Review this week, active days this month, and average workouts
                  per week with simple comparisons to the prior week.
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      <section className="px-6 py-16">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="mb-6 text-4xl font-bold leading-snug tracking-wide text-foreground md:text-6xl">
            Keep training data
            <br />
            <span className="text-muted-foreground">clear and reusable.</span>
          </h2>
          <p className="mb-8 text-xl leading-relaxed text-muted-foreground">
            FitTrack focuses on quick logging, visible progress, useful workout
            history, and practical consistency tracking.
          </p>

          <div className="mb-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button
              size="lg"
              className="bg-primary px-8 py-4 text-primary-foreground hover:bg-primary/90"
              asChild
            >
              <Link to="/workouts" preload={false}>
                <ArrowRight className="mr-2 h-5 w-5" />
                {user ? 'Open FitTrack' : 'Try FitTrack'}
              </Link>
            </Button>
            <Button
              size="lg"
              variant="outline"
              className="border-border bg-transparent px-8 py-4 text-muted-foreground hover:bg-background/50"
              asChild
            >
              <Link to="/workouts" preload={false}>
                See Workouts
              </Link>
            </Button>
          </div>

          <div className="text-sm text-muted-foreground">
            Track what happened. Reuse what worked. See where training is going.
          </div>
        </div>
      </section>
    </div>
  );
}

export const Route = createFileRoute('/')({
  component: App,
});

function App() {
  const { user } = Route.useRouteContext();
  return <HomePage user={user} />;
}
