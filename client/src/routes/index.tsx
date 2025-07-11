import { createFileRoute, Link } from '@tanstack/react-router';
import { useState } from 'react';
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

function HomePage({
  user,
}: {
  user: CurrentUser | CurrentInternalUser | null;
}) {
  const [selectedTestimonial, setSelectedTestimonial] = useState(null);

  // MARK: Features
  const features = [
    {
      icon: Zap,
      title: 'Lightning Fast',
      description:
        'Always responsive and efficient. Never compromises operational speed.',
      highlight: 'TACTICAL SPEED',
    },
    {
      icon: Target,
      title: 'Precision Interface',
      description:
        'Large touch zones optimized for field conditions and tactical gear.',
      highlight: 'FIELD READY',
    },
    {
      icon: Eye,
      title: 'High Visibility',
      description:
        'Maximum readability in all lighting conditions, from bunker to daylight.',
      highlight: 'ALL CONDITIONS',
    },
    {
      icon: Wifi,
      title: 'Offline Operations',
      description:
        'Full functionality without network dependency. Mission critical reliability.',
      highlight: 'ZERO DEPENDENCY',
    },
  ];

  // MARK: Testimonials
  const testimonials = [
    {
      id: 1,
      name: 'Agent Phoenix',
      role: 'Field Operative',
      rating: 5,
      text: "An elite and powerful tactical system. I've deployed many mission trackers over the years and this one is by far the best. The interface doesn't get in your way, and the design is so intuitive to work with. Best of all, the data are responsive and ultra quick to access even when bugs as they come up!",
      date: 'Nov 2023',
    },
    {
      id: 2,
      name: 'Commander Steel',
      role: 'Operations Director',
      rating: 5,
      text: 'Outstanding system and service. Command HQ and our field teams have recommended it to operatives of all different training levels and they have found it useful and intuitive. When I thought being able to sync data in real-time was just a dream, this system made it reality and notified me this was a feature in the works. Within 2 weeks of the communication the live export feature was live.',
      date: 'Oct 2023',
    },
    {
      id: 3,
      name: 'Operative Ghost',
      role: 'Tactical Specialist',
      rating: 5,
      text: "Mission critical! Very refreshing to see a tactical system that keeps mission logging quick and simple. Love the minimalistic design keeping only the essential functions in mind. I think linking on Command ID should be optional as it's not a necessity for keeping info on our local device (unless some would prefer a cloud backup). Looks like the system is still in its early days with the mission catalogue, but I'm excited to see where this system goes. Nice work!",
      date: 'Sep 2023',
    },
    {
      id: 4,
      name: 'Agent Viper',
      role: 'Intelligence Analyst',
      rating: 5,
      text: "Best tactical tracker. I started working out beginning of this year and was writing all my progress down in notebooks and on the notes app on my phone. Had this system pop up as an ad on instagram and I'm so glad I downloaded it. Its so easy to use, so clearly laid out, the team are constantly interacting with people and taking onboard feedback and its great to see. I use this every time I go to the gym and have recommended it to others because I think its so damn good.",
      date: 'Aug 2023',
    },
    {
      id: 5,
      name: 'Commander Raven',
      role: 'Strategic Operations',
      rating: 5,
      text: "Exceptional attention to detail and simplicity of the system. There's no bloat, easy to understand (unlike most tactical systems I've tried). Definitely sticking with this one. Incredible work.",
      date: 'Jul 2023',
    },
    {
      id: 6,
      name: 'Specialist Hawk',
      role: 'Field Commander',
      rating: 5,
      text: "Best Tracker I've Used. Has everything you need: all exercises, equipment, tracks weight, reps, etc. Future suggestion is to add a social aspect where you can share workouts with friends, kind of like the Apple Watch app.",
      date: 'Jun 2023',
    },
  ];

  return (
    <div className="min-h-screen bg-black text-white">
      {/* MARK: Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-black/90 backdrop-blur-sm border-b border-neutral-800">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="w-6 h-6 text-orange-500" />
            <span className="text-xl font-bold tracking-wider">FITTRACK</span>
            {user && (
              <>
                <div className="px-2 font-bold">
                  <Link to="/workouts/new">New Workout</Link>
                </div>
                <div className="px-2 font-bold">
                  <Link to="/workouts">Workouts</Link>
                </div>
                <div className="px-2 font-bold">
                  <Link to="/exercises">Exercises</Link>
                </div>
              </>
            )}
          </div>
          <div className="flex items-center gap-4">
            <Button className="bg-orange-500 hover:bg-orange-600 text-white">
              <Download className="w-4 h-4 mr-2" />
              Deploy System
            </Button>
            <UserButton />
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="pt-24 pb-16 px-6">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-12">
            <Badge className="bg-orange-500/20 text-orange-500 mb-6 px-4 py-2">
              <Shield className="w-4 h-4 mr-2" />
              Elite Operations Platform
            </Badge>
            <h1 className="text-5xl md:text-7xl font-bold tracking-wider mb-6">
              Serious tracking
              <br />
              for serious <span className="text-orange-500">fitness</span>
              <br />
              operations.
            </h1>
            <div className="flex items-center justify-center gap-4 mb-8">
              <Badge className="bg-neutral-800 text-white px-4 py-2">
                <Star className="w-4 h-4 mr-2 text-orange-500" />
                Elite Systems Award 2024
              </Badge>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <div className="space-y-8">
              <div className="space-y-4">
                <h3 className="text-xl font-medium text-neutral-300">
                  Mission Control
                </h3>
                <p className="text-neutral-400 leading-relaxed">
                  Advanced fitness operations management system designed for
                  elite athletes. Track workouts, manage training programs, and
                  coordinate fitness activities with military-grade precision
                  and reliability.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-4 bg-neutral-900 border border-neutral-700 rounded">
                  <div className="text-2xl font-bold text-orange-500 font-mono">
                    847
                  </div>
                  <div className="text-xs text-neutral-400">ACTIVE AGENTS</div>
                </div>
                <div className="text-center p-4 bg-neutral-900 border border-neutral-700 rounded">
                  <div className="text-2xl font-bold text-white font-mono">
                    23
                  </div>
                  <div className="text-xs text-neutral-400">ONGOING OPS</div>
                </div>
              </div>
            </div>

            <div className="flex justify-center">
              <div className="relative">
                <div className="w-80 h-96 bg-neutral-900 border border-neutral-700 rounded-3xl p-6 shadow-2xl">
                  <div className="text-center mb-6">
                    <div className="text-xs text-neutral-400 mb-2">
                      TACTICAL COMMAND
                    </div>
                    <div className="text-2xl font-bold text-white font-mono">
                      4:44
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div className="bg-neutral-800 border border-neutral-700 rounded p-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-white">
                          Mission Alpha
                        </span>
                        <Badge className="bg-orange-500/20 text-orange-500 text-xs">
                          ACTIVE
                        </Badge>
                      </div>
                      <div className="text-xs text-neutral-400">
                        190 lbs • 5 reps
                      </div>
                    </div>

                    <div className="bg-neutral-800 border border-neutral-700 rounded p-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-white">Recon Beta</span>
                        <Badge className="bg-white/20 text-white text-xs">
                          STANDBY
                        </Badge>
                      </div>
                      <div className="text-xs text-neutral-400">
                        Body weight • 21 min ago
                      </div>
                    </div>

                    <div className="bg-neutral-800 border border-neutral-700 rounded p-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-white">
                          Extraction Gamma
                        </span>
                        <Badge className="bg-white/20 text-white text-xs">
                          COMPLETE
                        </Badge>
                      </div>
                      <div className="text-xs text-neutral-400">
                        3,200 lbs total volume
                      </div>
                    </div>
                  </div>
                </div>

                {/* Floating notification */}
                <div className="absolute -top-4 -right-4 bg-orange-500 text-white rounded-full w-8 h-8 flex items-center justify-center text-sm font-bold">
                  1
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 px-6 bg-neutral-950">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-6xl font-bold tracking-wider mb-4">
              Tactical
              <br />
              <span className="text-neutral-500">by design.</span>
            </h2>
            <p className="text-xl text-neutral-400">
              Built exclusively for operational excellence.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {features.map((feature, index) => (
              <Card
                key={index}
                className="bg-neutral-900 border-neutral-700 p-6"
              >
                <CardContent className="p-0">
                  <div className="flex items-start gap-4">
                    <div className="p-3 bg-orange-500/20 rounded">
                      <feature.icon className="w-6 h-6 text-orange-500" />
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <h3 className="text-lg font-bold text-white">
                          {feature.title}
                        </h3>
                        <Badge className="bg-orange-500/20 text-orange-500 text-xs">
                          {feature.highlight}
                        </Badge>
                      </div>
                      <p className="text-neutral-400">{feature.description}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Advanced Features */}
      <section className="py-16 px-6">
        <div className="max-w-7xl mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <div>
              <h2 className="text-4xl md:text-6xl font-bold tracking-wider mb-6">
                Not just
                <br />
                <span className="text-neutral-500">missions.</span>
              </h2>
              <p className="text-xl text-neutral-400 mb-8 leading-relaxed">
                Advanced coordination protocols, real-time intelligence sharing,
                multi-vector analysis, tactical planning tools, resource
                allocation - all integrated into one comprehensive command
                system designed for maximum operational effectiveness.
              </p>

              <div className="space-y-4 text-neutral-400">
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-orange-500" />
                  <span>Real-time coordination</span>
                </div>
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-orange-500" />
                  <span>Intelligence integration</span>
                </div>
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-orange-500" />
                  <span>Resource optimization</span>
                </div>
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-orange-500" />
                  <span>Mission analytics</span>
                </div>
              </div>
            </div>

            <div className="flex justify-center">
              <div className="w-80 h-96 bg-neutral-900 border border-neutral-700 rounded-3xl p-6 shadow-2xl">
                <div className="text-center mb-6">
                  <div className="text-xs text-neutral-400 mb-2">
                    OPERATION TIMER
                  </div>
                  <div className="text-4xl font-bold text-white font-mono">
                    32:08
                  </div>
                  <div className="text-xs text-neutral-400">
                    Mission in progress
                  </div>
                </div>

                <div className="space-y-3">
                  <div className="flex items-center justify-between p-2 bg-neutral-800 rounded">
                    <div className="flex items-center gap-2">
                      <Target className="w-4 h-4 text-orange-500" />
                      <span className="text-sm text-white">Infiltration</span>
                    </div>
                    <span className="text-xs text-neutral-400">
                      5 km • 2:30 min/km
                    </span>
                  </div>

                  <div className="flex items-center justify-between p-2 bg-neutral-800 rounded">
                    <div className="flex items-center gap-2">
                      <Activity className="w-4 h-4 text-white" />
                      <span className="text-sm text-white">Surveillance</span>
                    </div>
                    <span className="text-xs text-neutral-400">
                      12 targets completed
                    </span>
                  </div>

                  <div className="flex items-center justify-between p-2 bg-neutral-800 rounded">
                    <div className="flex items-center gap-2">
                      <Shield className="w-4 h-4 text-white" />
                      <span className="text-sm text-white">Security</span>
                    </div>
                    <span className="text-xs text-neutral-400">3 x 30 sec</span>
                  </div>

                  <div className="flex items-center justify-between p-2 bg-neutral-800 rounded">
                    <div className="flex items-center gap-2">
                      <BarChart3 className="w-4 h-4 text-white" />
                      <span className="text-sm text-white">Analysis</span>
                    </div>
                    <span className="text-xs text-neutral-400">
                      Real-time data
                    </span>
                  </div>

                  <Button className="w-full bg-orange-500 hover:bg-orange-600 text-white mt-4">
                    <ArrowRight className="w-4 h-4 mr-2" />
                    Next Phase
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section className="py-16 px-6 bg-neutral-950">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-6xl font-bold tracking-wider mb-4">
              What elite
              <br />
              operatives think.
            </h2>
            <p className="text-xl text-neutral-400">
              Reviews from our field agents.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-16">
            {testimonials.slice(0, 4).map((testimonial) => (
              <Card
                key={testimonial.id}
                className="bg-neutral-900 border-neutral-700 p-6"
              >
                <CardContent className="p-0">
                  <div className="flex items-center gap-1 mb-4">
                    {[...Array(5)].map((_, i) => (
                      <Star
                        key={i}
                        className="w-4 h-4 fill-orange-500 text-orange-500"
                      />
                    ))}
                  </div>
                  <h3 className="text-lg font-bold text-white mb-2">
                    {testimonial.name}
                  </h3>
                  <p className="text-sm text-neutral-400 mb-4">
                    {testimonial.role}
                  </p>
                  <p className="text-neutral-300 leading-relaxed mb-4">
                    {testimonial.text}
                  </p>
                  <div className="text-xs text-neutral-500">
                    {testimonial.date}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Video Testimonial */}
          <Card className="bg-neutral-900 border-neutral-700 overflow-hidden">
            <CardContent className="p-0">
              <div className="grid grid-cols-1 lg:grid-cols-2">
                <div className="p-8">
                  <h3 className="text-2xl font-bold text-white mb-2">
                    Agent Phantom on
                    <br />
                    using FitTrack.
                  </h3>
                  <p className="text-neutral-400 mb-6">
                    How he uses FitTrack to coordinate his training operations.
                  </p>
                </div>
                <div className="relative bg-neutral-800 aspect-video flex items-center justify-center">
                  <div className="w-64 h-64 bg-neutral-700 rounded-full flex items-center justify-center">
                    <Button
                      size="lg"
                      className="bg-white hover:bg-neutral-200 text-black rounded-full w-16 h-16"
                    >
                      <Play className="w-6 h-6 ml-1" />
                    </Button>
                  </div>
                  <div className="absolute inset-0 bg-gradient-to-t from-black/50 to-transparent" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-16 px-6">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl md:text-6xl font-bold tracking-wider mb-6">
            Deploy from
            <br />
            FitTrack HQ.
          </h2>
          <p className="text-xl text-neutral-400 mb-8 leading-relaxed">
            FitTrack is available for deployment and field testing. After
            initial assessment, full access requires fitness clearance and
            training approval. Designed by Elite Fitness Command, Stockholm.
            Advanced training systems since 2020. Contact Command with training
            requirements or fitness feedback at{' '}
            <span className="text-orange-500">command@fittrack.app</span>
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-8">
            <Button
              size="lg"
              className="bg-orange-500 hover:bg-orange-600 text-white px-8 py-4"
            >
              <Download className="w-5 h-5 mr-2" />
              Request Deployment
            </Button>
            <Button
              size="lg"
              variant="outline"
              className="border-neutral-700 text-neutral-300 hover:bg-neutral-800 px-8 py-4 bg-transparent"
            >
              View Documentation
            </Button>
          </div>

          <div className="text-sm text-neutral-500">
            Classification Level: CONFIDENTIAL • 2024
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
