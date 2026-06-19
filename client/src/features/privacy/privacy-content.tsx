import type { ReactNode } from "react";

export const privacyEmail = "privacy@fittrack.andrewy.me";
export const supportEmail = "support@fittrack.andrewy.me";
export const effectiveDate = "June 18, 2026";

export const dataCategories = [
  {
    title: "Account information",
    body: "FitTrack uses Stack Auth for sign-in. FitTrack stores an account identifier in its app database and may receive basic account details from Stack Auth, such as email, display name, and profile image, when those details are available through your account.",
  },
  {
    title: "Workout information",
    body: "FitTrack stores the weightlifting information you enter, including workout dates, notes, workout focus, exercise names, sets, reps, weight, set type, exercise order, and calculated training history such as estimated one-rep max values.",
  },
  {
    title: "AI chat information",
    body: "If you use AI chat, FitTrack stores chat messages, generated workout drafts, conversation state, and related usage details. Messages may include fitness goals, equipment, session length, training level, limitations, or injury-related context if you choose to provide that information.",
  },
  {
    title: "Billing information",
    body: "FitTrack uses Stripe for paid AI chat subscriptions. FitTrack stores Stripe customer and subscription identifiers, subscription status, trial usage, billing period dates, cancellation state, and webhook event records. FitTrack does not store full payment card numbers.",
  },
  {
    title: "Device, usage, and local storage information",
    body: "FitTrack may process technical details such as browser, device, request, and app usage information to run the service, secure accounts, troubleshoot issues, and understand product reliability. The app also uses browser storage for local drafts, demo data, theme preference, and app state.",
  },
];

export const useCases = [
  "Provide workout logging, exercise history, analytics, and AI-generated workout drafts.",
  "Create and maintain your account and keep your data associated with the right signed-in user.",
  "Process subscriptions, trials, cancellations, billing status, and customer portal access through Stripe.",
  "Secure the app, prevent misuse, debug errors, monitor reliability, and enforce per-user data access.",
  "Respond to support, privacy, deletion, and data access requests.",
  "Comply with billing, tax, security, dispute, and legal obligations.",
];

export const processors = [
  {
    name: "Stack Auth",
    purpose: "account authentication and sign-in",
  },
  {
    name: "Stripe",
    purpose:
      "subscription checkout, billing portal, payment processing, billing records, and subscription webhooks",
  },
  {
    name: "Google Gemini API",
    purpose: "AI chat responses and workout draft generation",
  },
  {
    name: "Supabase",
    purpose: "PostgreSQL database hosting",
  },
  {
    name: "Fly.io",
    purpose: "application hosting",
  },
  {
    name: "Inngest",
    purpose: "AI chat recovery and background workflow support, when enabled",
  },
];

export function PolicyLink({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}) {
  return (
    <a
      href={href}
      className="font-medium text-primary underline-offset-4 hover:underline"
    >
      {children}
    </a>
  );
}

export function PolicyList({ items }: { items: string[] }) {
  return (
    <ul className="space-y-3">
      {items.map((item) => (
        <li
          key={item}
          className="flex gap-3"
        >
          <span
            aria-hidden="true"
            className="mt-3 h-1.5 w-1.5 shrink-0 rounded-full bg-primary"
          />
          <span>{item}</span>
        </li>
      ))}
    </ul>
  );
}

/**
 * Shared section bodies. Each section is keyed by an id (used for TOC anchors in
 * the docs layout) and carries its title + rendered content. Keeping the copy
 * here means both privacy route variants stay in sync.
 */
export function getPolicySections(): {
  id: string;
  title: string;
  content: ReactNode;
}[] {
  return [
    {
      id: "who-we-are",
      title: "Who We Are",
      content: (
        <>
          <p>
            FitTrack is a weightlifting workout tracker that helps users log
            workouts, review exercise history, view training analytics, and
            generate workout drafts with AI chat.
          </p>
          <p>
            For privacy questions, contact{" "}
            <PolicyLink href={`mailto:${privacyEmail}`}>
              {privacyEmail}
            </PolicyLink>
            . For general support, contact{" "}
            <PolicyLink href={`mailto:${supportEmail}`}>
              {supportEmail}
            </PolicyLink>
            .
          </p>
        </>
      ),
    },
    {
      id: "information-we-collect",
      title: "Information We Collect",
      content: (
        <div className="grid gap-4">
          {dataCategories.map((category) => (
            <div
              key={category.title}
              className="rounded-xl border border-border bg-card p-5"
            >
              <h3 className="mb-2 font-semibold text-foreground">
                {category.title}
              </h3>
              <p>{category.body}</p>
            </div>
          ))}
        </div>
      ),
    },
    {
      id: "how-we-use-information",
      title: "How We Use Information",
      content: <PolicyList items={useCases} />,
    },
    {
      id: "payments",
      title: "Payments",
      content: (
        <>
          <p>
            FitTrack uses Stripe-hosted payment and billing flows for paid AI
            chat subscriptions, including checkout, subscription status,
            customer portal access, cancellation flows, and billing webhooks.
          </p>
          <p>
            Stripe processes payment details. FitTrack stores the billing
            identifiers and subscription information needed to provide paid
            access and troubleshoot billing issues, but FitTrack does not store
            full card numbers.
          </p>
          <p>
            Stripe may process information according to its own privacy terms.
            You can review Stripe's privacy policy at{" "}
            <PolicyLink href="https://stripe.com/privacy">
              stripe.com/privacy
            </PolicyLink>
            .
          </p>
        </>
      ),
    },
    {
      id: "ai-chat-and-fitness-information",
      title: "AI Chat And Fitness Information",
      content: (
        <>
          <p>
            FitTrack AI chat can use details you provide, such as goals,
            available equipment, workout focus, training level, session length,
            and limitations, to generate a workout draft. You should avoid
            entering information you do not want processed by the AI feature.
          </p>
          <p>
            FitTrack is for workout tracking and fitness planning. It is not a
            medical service, and AI-generated workouts are not medical advice.
          </p>
        </>
      ),
    },
    {
      id: "how-we-share-information",
      title: "How We Share Information",
      content: (
        <>
          <p>
            FitTrack does not sell personal information and does not share
            personal information for targeted advertising. FitTrack shares
            information with service providers that help operate the app:
          </p>
          <div className="grid gap-3">
            {processors.map((processor) => (
              <div
                key={processor.name}
                className="rounded-xl border border-border bg-card p-4"
              >
                <span className="font-semibold text-foreground">
                  {processor.name}
                </span>
                <span className="text-muted-foreground">
                  {" "}
                  - {processor.purpose}
                </span>
              </div>
            ))}
          </div>
          <p>
            FitTrack may also disclose information when required for legal,
            security, fraud prevention, billing, dispute, or business transfer
            purposes.
          </p>
        </>
      ),
    },
    {
      id: "your-choices-and-rights",
      title: "Your Choices And Rights",
      content: (
        <>
          <p>
            You can update or delete workouts, exercises, and other app data
            through FitTrack features where those controls are available. You
            can delete your FitTrack account in the app.
          </p>
          <p>
            If you want a copy of your data, request it at{" "}
            <PolicyLink href={`mailto:${privacyEmail}`}>
              {privacyEmail}
            </PolicyLink>{" "}
            before deleting your account. After deletion, FitTrack may no longer
            have the workout, AI chat, or account data needed to fulfill an
            access or export request.
          </p>
          <p>
            Depending on where you live, you may have rights to request access,
            correction, deletion, portability, or more information about how
            your information is used. You may also have the right to appeal a
            privacy request decision. Contact{" "}
            <PolicyLink href={`mailto:${privacyEmail}`}>
              {privacyEmail}
            </PolicyLink>{" "}
            to make a privacy request.
          </p>
        </>
      ),
    },
    {
      id: "california-and-us-state-privacy-rights",
      title: "California And U.S. State Privacy Rights",
      content: (
        <>
          <p>
            Some U.S. state privacy laws give residents additional rights over
            personal information. FitTrack will review requests based on the law
            that applies to your location and to FitTrack.
          </p>
          <p>
            California residents may have rights to know, access, correct,
            delete, and receive a copy of certain information, and to opt out of
            sale or sharing where applicable. FitTrack does not sell personal
            information or share it for targeted advertising.
          </p>
        </>
      ),
    },
    {
      id: "consumer-health-data",
      title: "Consumer Health Data",
      content: (
        <>
          <p>
            FitTrack is a workout tracking app. Workout entries, exercise
            history, AI chat messages, fitness goals, limitations, and
            injury-related context you choose to provide may be considered
            health, wellness, or consumer health data under some laws.
          </p>
          <p>
            FitTrack uses this information to provide workout tracking,
            analytics, AI chat, account support, security, and app operations.
            FitTrack does not sell this information.
          </p>
        </>
      ),
    },
    {
      id: "children",
      title: "Children",
      content: (
        <p>
          FitTrack is not intended for children under 13, and we do not
          knowingly collect information from children under 13. If you believe a
          child has provided information to FitTrack, contact{" "}
          <PolicyLink href={`mailto:${privacyEmail}`}>
            {privacyEmail}
          </PolicyLink>
          .
        </p>
      ),
    },
    {
      id: "retention",
      title: "Retention",
      content: (
        <>
          <p>
            FitTrack keeps account, workout, AI chat, and billing information
            while needed to provide the service. When you delete your account,
            FitTrack deletes or de-identifies FitTrack app data unless limited
            information must be kept for billing, security, legal, fraud
            prevention, or dispute purposes.
          </p>
          <p>
            Some information may remain temporarily in logs or provider systems
            where needed for security, billing, legal, or operations purposes.
          </p>
        </>
      ),
    },
    {
      id: "security",
      title: "Security",
      content: (
        <p>
          FitTrack uses safeguards designed to protect FitTrack data, including
          authenticated access, per-user data boundaries, HTTPS in production,
          Stripe webhook signature verification, rate limiting, and limited
          access to production systems.
        </p>
      ),
    },
    {
      id: "changes-to-this-policy",
      title: "Changes To This Policy",
      content: (
        <p>
          FitTrack may update this Privacy Policy as the app, vendors, or legal
          requirements change. The effective date at the top of the page shows
          when this policy was last updated.
        </p>
      ),
    },
  ];
}
