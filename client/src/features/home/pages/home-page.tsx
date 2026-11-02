import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { ArrowRight, Bot, Dumbbell, Target } from "lucide-react";
import type { ApplicationUser } from "@/lib/application-user";
import { AppBottomBar } from "@/components/nav/app-bottom-bar";
import { LandingTopBar } from "@/components/nav/landing-top-bar";
import { useDisplayMode } from "@/hooks/use-display-mode";
import { HowItWorks } from "@/features/home/components/how-it-works";

export function HomePage({ user }: { user: ApplicationUser | null }) {
  const displayMode = useDisplayMode();

  const showPwaGuestBottomBar = displayMode === "pwa" && !user;

  return (
    <div
      className={
        showPwaGuestBottomBar
          ? "min-h-screen pb-[calc(5rem+env(safe-area-inset-bottom))]"
          : "min-h-screen"
      }
    >
      <LandingTopBar user={user} />
      {showPwaGuestBottomBar ? <AppBottomBar user={user} /> : null}

      <section className="px-6 pb-16 pt-24">
        <div className="@container mx-auto max-w-7xl">
          <div className="space-y-6 text-center">
            {/* Use the content width, excluding padding and visible scrollbars. */}
            <h1 className="text-[length:clamp(1.5rem,calc(100cqw/10.2),4.5rem)] font-bold leading-snug tracking-wide text-foreground">
              Log workouts fast.
              <br />
              See progress clearly.
            </h1>
            <p className="mx-auto max-w-3xl text-lg leading-relaxed text-muted-foreground">
              Track exercises, sets, reps, weight, and notes, then review your
              history and progress in one place.
            </p>
            <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
              <Button
                size="lg"
                className="bg-primary px-8 py-4 text-primary-foreground hover:bg-primary/90"
                asChild
              >
                <Link
                  to="/workouts"
                  preload={false}
                >
                  <Dumbbell className="mr-1 h-5 w-5" />
                  {user ? "Open Workouts" : "Try the App"}
                </Link>
              </Button>
              <Button
                size="lg"
                variant="outline"
                className="border-border bg-transparent px-8 py-4 text-muted-foreground hover:bg-background/50"
                asChild
              >
                <Link
                  to="/exercises"
                  preload={false}
                >
                  Browse Exercise History
                </Link>
              </Button>
            </div>
            {user ? null : (
              <p className="text-sm text-muted-foreground">
                No account needed to try.
              </p>
            )}
          </div>
        </div>
      </section>

      <HowItWorks />

      <section
        aria-labelledby="also-in-fittrack-heading"
        className="px-6 py-16"
      >
        <div className="mx-auto max-w-4xl">
          <h2
            id="also-in-fittrack-heading"
            className="mb-8 text-center text-sm font-semibold uppercase tracking-widest text-primary"
          >
            Also in FitTrack
          </h2>
          <div className="grid gap-8 md:grid-cols-2">
            <div>
              <h3 className="flex items-center gap-3 text-lg font-semibold text-foreground">
                <Target className="h-5 w-5 shrink-0 text-primary" />
                Exercise goals
              </h3>
              <p className="mt-2 leading-7 text-muted-foreground">
                Set a target weight, reps, or weekly frequency for any exercise.
              </p>
            </div>
            <div>
              <h3 className="flex items-center gap-3 text-lg font-semibold text-foreground">
                <Bot className="h-5 w-5 shrink-0 text-primary" />
                AI Chat
              </h3>
              <p className="mt-2 leading-7 text-muted-foreground">
                Ask about your training and get workout drafts you can save.
                Requires sign-in; paid after a 7-day trial of 30 prompts.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="px-6 py-16">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="mb-8 text-[length:clamp(1.5rem,calc((100vw-3rem)/11.5),3.75rem)] font-bold leading-snug tracking-wide text-foreground">
            Keep training data
            <br />
            <span className="text-muted-foreground">clear and reusable.</span>
          </h2>
          <Button
            size="lg"
            className="bg-primary px-8 py-4 text-primary-foreground hover:bg-primary/90"
            asChild
          >
            <Link
              to="/workouts"
              preload={false}
            >
              <ArrowRight className="mr-2 h-5 w-5" />
              {user ? "Open FitTrack" : "Try FitTrack"}
            </Link>
          </Button>
        </div>
      </section>

      <footer className="border-t border-border px-6 py-8">
        <div className="mx-auto flex max-w-7xl flex-col gap-3 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
          <span>FitTrack</span>
          <div className="flex gap-4">
            <Link
              to="/privacy"
              className="hover:text-foreground"
            >
              Privacy Policy
            </Link>
            <a
              href="mailto:support@fittrack.andrewy.me"
              className="hover:text-foreground"
            >
              Support
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
