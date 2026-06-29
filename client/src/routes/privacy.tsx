import { createFileRoute } from "@tanstack/react-router";
import { ShieldCheck } from "lucide-react";
import { AppShell } from "@/components/nav/app-shell";
import { Badge } from "@/components/ui/badge";
import {
  getPolicySections,
  effectiveDate,
} from "@/features/privacy/privacy-content";
import { useDisplayMode } from "@/hooks/use-display-mode";

export const Route = createFileRoute("/privacy")({
  component: PrivacyPage,
});

export function PrivacyPage() {
  const { user } = Route.useRouteContext();
  const displayMode = useDisplayMode();
  const sections = getPolicySections();

  return (
    <main
      className={
        displayMode === "pwa"
          ? "min-h-screen bg-background pt-[env(safe-area-inset-top)] pb-[calc(5rem+env(safe-area-inset-bottom))]"
          : "min-h-screen bg-background"
      }
    >
      <AppShell user={user} />
      <section className="px-6 py-14">
        <div className="mx-auto max-w-6xl">
          <div className="mb-12 max-w-3xl space-y-4">
            <Badge className="bg-primary/15 px-4 py-2 text-primary">
              <ShieldCheck className="mr-1 h-4 w-4" />
              Effective {effectiveDate}
            </Badge>
            <h1 className="text-4xl font-bold leading-tight tracking-wide text-foreground md:text-6xl">
              Privacy Policy
            </h1>
            <p className="text-lg leading-8 text-muted-foreground">
              This Privacy Policy explains how FitTrack collects, uses, shares,
              and protects information when you use the FitTrack website, app,
              workout tracking features, AI chat, and billing flows.
            </p>
          </div>

          <div className="grid gap-10 lg:grid-cols-[16rem_1fr] lg:gap-16">
            <aside className="hidden lg:block">
              <nav className="sticky top-8">
                <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  On this page
                </p>
                <ul className="space-y-1 border-l border-border">
                  {sections.map((section) => (
                    <li key={section.id}>
                      <a
                        href={`#${section.id}`}
                        className="-ml-px block border-l border-transparent py-1.5 pl-4 text-sm text-muted-foreground transition-colors hover:border-primary hover:text-foreground"
                      >
                        {section.title}
                      </a>
                    </li>
                  ))}
                </ul>
              </nav>
            </aside>

            <div className="min-w-0 space-y-10">
              {sections.map((section) => (
                <section
                  key={section.id}
                  id={section.id}
                  className="scroll-mt-8 border-t border-border pt-8 first:border-t-0 first:pt-0"
                >
                  <h2 className="mb-4 text-2xl font-bold tracking-wide text-foreground">
                    {section.title}
                  </h2>
                  <div className="space-y-4 leading-7 text-muted-foreground">
                    {section.content}
                  </div>
                </section>
              ))}
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
