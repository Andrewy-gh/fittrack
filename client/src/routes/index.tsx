import { createFileRoute, Link } from '@tanstack/react-router';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Shield,
  Target,
  Zap,
  Eye,
  Wifi,
  Download,
  Star,
  Play,
  ArrowRight,
  CheckCircle,
  Activity,
  BarChart3,
} from 'lucide-react';
import { UserButton } from '@stackframe/react';
import { stackClientApp } from '@/stack';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import { ModeToggle } from '@/components/mode-toggle';

function HomePage({
  user,
}: {
  user: CurrentUser | CurrentInternalUser | null;
}) {
  // MARK: Features
  const features = [
    {
      icon: Zap,
      title: 'Never Miss a Rep',
      description:
        'Log workouts in seconds, not minutes. Our streamlined interface captures every set while you focus on crushing your goals.',
      highlight: 'LIGHTNING FAST',
    },
    {
      icon: Target,
      title: 'Smart Progress Tracking',
      description:
        'See exactly how each workout builds toward your bigger goals. Visual progress that keeps you motivated to push harder.',
      highlight: 'RESULTS FOCUSED',
    },
    {
      icon: Eye,
      title: 'Your Wins Made Visible',
      description:
        'Transform scattered workouts into clear victories. Beautiful charts show your strength gains and consistency streaks.',
      highlight: 'SEE YOUR SUCCESS',
    },
    {
      icon: Wifi,
      title: 'Works Everywhere',
      description:
        'Gym, home, or outdoors - your progress syncs seamlessly. Never lose a workout, even when offline.',
      highlight: 'ALWAYS READY',
    },
  ];

  // MARK: Testimonials
  const testimonials = [
    {
      id: 1,
      name: 'Alex Strong',
      role: 'Fitness Enthusiast',
      rating: 5,
      text: "Finally, a fitness app that doesn't get in my way. FitTrack makes logging workouts so fast I actually stick with it. My consistency has never been better.",
      date: 'July 2025',
    },
    {
      id: 2,
      name: 'Jordan Peak',
      role: 'Personal Trainer',
      rating: 5,
      text: "I recommend FitTrack to all my clients. It's the only app that makes tracking feel effortless while showing real progress. Game changer for habit building.",
      date: 'June 2025',
    },
    {
      id: 3,
      name: 'Taylor Lift',
      role: 'Powerlifter',
      rating: 5,
      text: 'The progress visualization is incredible. Seeing my strength gains mapped out keeps me motivated through tough training blocks. Worth every download.',
      date: 'May 2025',
    },
    {
      id: 4,
      name: 'Chris Endurance',
      role: 'Runner',
      rating: 5,
      text: "Love how FitTrack is expanding beyond weightlifting. The foundation is rock solid, and I'm excited to track my runs here soon!",
      date: 'April 2025',
    },
  ];

  return (
    <div className="min-h-screen">
      {/* MARK: Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-background/90 backdrop-blur-sm border-b border-border">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="w-6 h-6 text-primary" />
            <span className="text-xl font-bold tracking-wide text-foreground">
              FITTRACK
            </span>
            {user && (
              <>
                <div className="px-2 font-bold text-foreground">
                  <Link to="/workouts">Workouts</Link>
                </div>
                <div className="px-2 font-bold text-foreground">
                  <Link to="/exercises">Exercises</Link>
                </div>
              </>
            )}
          </div>
          <div className="flex items-center gap-4">
            <ModeToggle />
            <UserButton />
            <Button className="bg-primary text-primary-foreground hover:bg-primary/90">
              <Download className="w-4 h-4 mr-2" />
              Get Started Free
            </Button>
          </div>
        </div>
      </nav>

      {/* MARK: Hero Section */}
      <section className="pt-24 pb-16 px-6">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-12 space-y-6">
            <Badge className="bg-primary/20 text-primary mb-6 px-4 py-2">
              <Shield className="w-4 h-4 mr-2 fill-primary" />
              Turn Fitness Into Your Daily Win
            </Badge>
            <h1 className="mb-6 text-5xl leading-snug font-bold tracking-wide text-foreground md:text-7xl">
              Build Your Strongest
              <br />
              <span className="text-primary">Habit</span> Yet
              <br />
              With FitTrack.
            </h1>
            <p className="text-muted-foreground leading-relaxed">
              Stop starting over. FitTrack transforms scattered workouts into
              lasting habits through smart tracking, visual progress, and a
              community that keeps you accountable. Finally see the results
              you're working for.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-8">
              <Button
                size="lg"
                className="bg-primary text-primary-foreground hover:bg-primary/90 px-8 py-4"
              >
                <Download className="w-5 h-5 mr-2" />
                Download Free Today
              </Button>
              <Button
                size="lg"
                variant="outline"
                className="border-border text-muted-foreground hover:bg-background/50 px-8 py-4 bg-transparent"
              >
                See How It Works
              </Button>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <div className="space-y-8">
              <div className="space-y-4">
                <h3 className="text-xl font-medium text-muted-foreground">
                  Your Progress, Simplified
                </h3>
                <p className="text-muted-foreground leading-relaxed">
                  FitTrack makes every workout count by showing you exactly how
                  each session builds toward your bigger goals. No more guessing
                  if you're making progress.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-4 bg-card border border-border rounded space-y-3">
                  <div className="text-2xl font-bold text-primary">
                    250K+
                  </div>
                  <div className="text-xs text-muted-foreground">
                    ACTIVE USERS
                  </div>
                </div>
                <div className="text-center p-4 bg-card border border-border rounded space-y-3">
                  <div className="text-2xl font-bold text-foreground">
                    85%
                  </div>
                  <div className="text-xs text-muted-foreground">
                    STILL ACTIVE AFTER 6 MONTHS
                  </div>
                </div>
              </div>
            </div>

            <div className="flex justify-center">
              <div className="relative">
                <div className="w-80 h-96 bg-card border border-border rounded-3xl p-6 shadow-2xl">
                  <div className="text-center mb-6">
                    <div className="text-xs text-muted-foreground mb-2">
                      CURRENT WORKOUT
                    </div>
                    <div className="text-2xl font-bold text-foreground">
                      30:15
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div className="bg-card border border-border rounded p-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-foreground">
                          Bench Press
                        </span>
                        <Badge className="bg-primary/20 text-primary text-xs">
                          CRUSHING IT
                        </Badge>
                      </div>
                      <div className="text-xs text-muted-foreground">
                        200 lbs • 8 reps • +5 lbs from last week
                      </div>
                    </div>

                    <div className="bg-card border border-border rounded p-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-foreground">Squats</span>
                        <Badge className="bg-secondary/20 text-secondary-foreground text-xs">
                          COMPLETE
                        </Badge>
                      </div>
                      <div className="text-xs text-muted-foreground">
                        300 lbs • 10 reps • New PR! 🎉
                      </div>
                    </div>

                    <Button className="w-full bg-primary text-primary-foreground hover:bg-primary/90 mt-4">
                      <ArrowRight className="w-4 h-4 mr-2" />
                      Log Next Set
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* MARK: Features Section */}
      <section className="py-16 px-6 bg-background">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="mb-4 text-4xl leading-snug font-bold tracking-wide text-foreground md:text-6xl">
              Simple
              <br />
              <span className="text-muted-foreground">by design.</span>
            </h2>
            <p className="text-xl text-muted-foreground">
              Built for people who want results, not complexity.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {features.map((feature, index) => (
              <Card key={index} className="bg-card border border-border p-6">
                <CardContent className="p-0">
                  <div className="flex items-start gap-4">
                    <div className="p-3 bg-primary/20 rounded">
                      <feature.icon className="w-6 h-6 text-primary" />
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <h3 className="text-lg font-bold text-foreground">
                          {feature.title}
                        </h3>
                        <Badge className="bg-primary/20 text-primary text-xs">
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

      {/* MARK: Advanced Features */}
      <section className="py-16 px-6">
        <div className="max-w-7xl mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <div>
              <h2 className="mb-6 text-4xl leading-snug font-bold tracking-wide text-foreground md:text-6xl">
                More than
                <br />
                <span className="text-muted-foreground">just logging.</span>
              </h2>
              <p className="text-xl text-muted-foreground mb-8 leading-relaxed">
                Smart goal setting that evolves with you. Community challenges
                that turn solo workouts into shared victories. Nutrition
                insights that fuel your performance. Everything you need to
                build the fitness habit that actually sticks.
              </p>

              <div className="space-y-4 text-muted-foreground">
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-primary" />
                  <span>AI-powered goal recommendations</span>
                </div>
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-primary" />
                  <span>Weekly community challenges</span>
                </div>
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-primary" />
                  <span>Integrated nutrition tracking</span>
                </div>
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-primary" />
                  <span>Habit streak visualization</span>
                </div>
              </div>
            </div>

            <div className="flex justify-center">
              <div className="w-80 h-96 bg-card border border-border rounded-3xl p-6 shadow-2xl">
                <div className="text-center mb-6">
                  <div className="text-xs text-muted-foreground mb-2">
                    THIS WEEK'S CHALLENGE
                  </div>
                  <div className="text-4xl font-bold text-foreground">
                    7/10
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Workouts completed
                  </div>
                </div>

                <div className="space-y-3">
                  <div className="flex items-center justify-between p-2 bg-card border border-border rounded">
                    <div className="flex items-center gap-2">
                      <Target className="w-4 h-4 text-primary" />
                      <span className="text-sm text-foreground">
                        Upper Body
                      </span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      45 min • Complete ✓
                    </span>
                  </div>

                  <div className="flex items-center justify-between p-2 bg-card border border-border rounded">
                    <div className="flex items-center gap-2">
                      <Activity className="w-4 h-4 text-foreground" />
                      <span className="text-sm text-foreground">
                        Lower Body
                      </span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      50 min • Complete ✓
                    </span>
                  </div>

                  <div className="flex items-center justify-between p-2 bg-card border border-border rounded">
                    <div className="flex items-center gap-2">
                      <Shield className="w-4 h-4 text-foreground" />
                      <span className="text-sm text-foreground">Cardio</span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      30 min • Complete ✓
                    </span>
                  </div>

                  <div className="flex items-center justify-between p-2 bg-card border border-border rounded">
                    <div className="flex items-center gap-2">
                      <BarChart3 className="w-4 h-4 text-primary" />
                      <span className="text-sm text-foreground">Rest Day</span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      Recovery earned
                    </span>
                  </div>

                  <Button className="w-full bg-primary text-primary-foreground hover:bg-primary/90 mt-4">
                    <ArrowRight className="w-4 h-4 mr-2" />
                    Join Challenge
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* MARK: Testimonials */}
      <section className="py-16 px-6 bg-background">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="mb-4 text-4xl leading-snug font-bold tracking-wide text-foreground md:text-6xl">
              Real people,
              <br />
              <span className="text-muted-foreground">real results.</span>
            </h2>
            <p className="text-xl text-muted-foreground">
              Join 250,000+ people building their strongest habits yet.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-16">
            {testimonials.slice(0, 4).map((testimonial) => (
              <Card
                key={testimonial.id}
                className="bg-card border border-border p-6"
              >
                <CardContent className="p-0">
                  <div className="flex items-center gap-1 mb-4">
                    {[...Array(5)].map((_, i) => (
                      <Star
                        key={i}
                        className="w-4 h-4 fill-primary text-primary"
                      />
                    ))}
                  </div>
                  <h3 className="text-lg font-bold text-foreground mb-2">
                    {testimonial.name}
                  </h3>
                  <p className="text-sm text-muted-foreground mb-4">
                    {testimonial.role}
                  </p>
                  <p className="text-muted-foreground leading-relaxed mb-4">
                    "{testimonial.text}"
                  </p>
                  <div className="text-xs text-muted-foreground">
                    {testimonial.date}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* MARK: Video Testimonial */}
          <Card className="bg-card border border-border overflow-hidden">
            <CardContent className="p-0">
              <div className="grid grid-cols-1 lg:grid-cols-2">
                <div className="p-8">
                  <h3 className="text-2xl font-bold text-foreground mb-2">
                    "FitTrack changed
                    <br />
                    how I see progress."
                  </h3>
                  <p className="text-muted-foreground mb-6">
                    Watch Alex share how FitTrack helped her build a 6-month
                    workout streak and hit her strength goals.
                  </p>
                </div>
                <div className="relative bg-card aspect-video flex items-center justify-center">
                  <div className="w-64 h-64 bg-card rounded-full flex items-center justify-center">
                    <Button
                      size="lg"
                      className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-full w-16 h-16"
                    >
                      <Play className="w-6 h-6 ml-1" />
                    </Button>
                  </div>
                  <div className="absolute inset-0 bg-gradient-to-t from-background/50 to-transparent" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* MARK: CTA Section */}
      <section className="py-16 px-6">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="mb-6 text-4xl leading-snug font-bold tracking-wide text-foreground md:text-6xl">
            Ready to build your
            <br />
            <span className="text-muted-foreground">strongest habit yet?</span>
          </h2>
          <p className="text-xl text-muted-foreground mb-8 leading-relaxed">
            Download FitTrack free and join thousands who've transformed
            scattered workouts into consistent progress. Your future self will
            thank you.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-8">
            <Button
              size="lg"
              className="bg-primary text-primary-foreground hover:bg-primary/90 px-8 py-4"
            >
              <Download className="w-5 h-5 mr-2" />
              Download Free
            </Button>
            <Button
              size="lg"
              variant="outline"
              className="border-border text-muted-foreground hover:bg-background/50 px-8 py-4 bg-transparent"
            >
              See Features
            </Button>
          </div>

          <div className="text-sm text-muted-foreground">
            ✓ Free download • ✓ No credit card required • ✓ Works on iPhone &
            Android
          </div>
        </div>
      </section>
    </div>
  );
}

export const Route = createFileRoute('/')({
  loader: async () => {
    const user = await stackClientApp.getUser();
    if (!user) {
      return null;
    }
    return user;
  },
  component: App,
});

function App() {
  const user = Route.useLoaderData();
  return <HomePage user={user} />;
}
