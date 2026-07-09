import {
  useEffect,
  useMemo,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Save, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  trainingProfileQueryOptions,
  type TrainingProfile,
  type UpdateTrainingProfileRequest,
  useUpdateTrainingProfileMutation,
} from "@/features/training-profile/api/training-profile";
import { getErrorMessage, isApiError } from "@/lib/errors";

type SelectOption = {
  value: string;
  label: string;
};

type MovementState = "unspecified" | "none" | "list";

type FormState = {
  primaryGoal: string;
  experienceLevel: string;
  preferredSessionDurationMinutes: string;
  usualTrainingLocation: string;
  availableEquipment: string[];
  avoidedExercises: string[];
  movementState: MovementState;
  movementLimitations: string[];
};

type FieldErrors = Partial<Record<keyof FormState | "form", string>>;

const introCopy =
  "FitTrack's AI coach uses this profile to personalize workouts — it fills in your usual setup so you don't repeat yourself in chat. The AI updates it when you tell it something lasting; you can correct anything here.";

const movementHelperText =
  "If you choose 'No known limitations', the AI coach will stop asking about injuries before workouts. Leave 'Not specified' if you'd rather be asked.";

const movementLimitationsRequiredMessage =
  "Add at least one limitation, or choose another option.";

const goalOptions: SelectOption[] = [
  { value: "strength", label: "Strength" },
  { value: "hypertrophy", label: "Hypertrophy" },
  { value: "endurance", label: "Endurance" },
  { value: "general_fitness", label: "General fitness" },
  { value: "weight_loss", label: "Weight loss" },
  { value: "mobility", label: "Mobility" },
];

const experienceOptions: SelectOption[] = [
  { value: "beginner", label: "Beginner" },
  { value: "intermediate", label: "Intermediate" },
  { value: "advanced", label: "Advanced" },
];

const locationOptions: SelectOption[] = [
  { value: "gym", label: "Gym" },
  { value: "home", label: "Home" },
  { value: "outdoor", label: "Outdoor" },
  { value: "travel", label: "Travel" },
];

const emptyProfile: TrainingProfile = {
  primary_goal: null,
  experience_level: null,
  preferred_session_duration_minutes: null,
  usual_training_location: null,
  available_equipment: [],
  avoided_exercises: [],
  movement_limitations: null,
};

export function TrainingProfilePage() {
  const profileQuery = useQuery(trainingProfileQueryOptions());
  const mutation = useUpdateTrainingProfileMutation();
  const [form, setForm] = useState<FormState>(() =>
    formStateFromProfile(emptyProfile),
  );
  const [savedForm, setSavedForm] = useState<FormState>(() =>
    formStateFromProfile(emptyProfile),
  );
  const [errors, setErrors] = useState<FieldErrors>({});

  useEffect(() => {
    if (!profileQuery.data) {
      return;
    }
    const next = formStateFromProfile(profileQuery.data);
    setForm(next);
    setSavedForm(next);
    setErrors({});
  }, [profileQuery.data]);

  const payload = useMemo(() => payloadFromForm(form), [form]);
  const isDirty = useMemo(
    () => JSON.stringify(form) !== JSON.stringify(savedForm),
    [form, savedForm],
  );

  function updateForm(patch: Partial<FormState>) {
    setForm((current) => ({ ...current, ...patch }));
    setErrors({});
  }

  function validateForm(): FieldErrors {
    const nextErrors: FieldErrors = {};
    const durationText = form.preferredSessionDurationMinutes.trim();
    if (durationText !== "") {
      const duration = Number(durationText);
      if (!Number.isInteger(duration) || duration < 10 || duration > 240) {
        nextErrors.preferredSessionDurationMinutes =
          "Enter a duration from 10 to 240 minutes.";
      }
    }
    if (
      form.movementState === "list" &&
      cleanTagList(form.movementLimitations).length === 0
    ) {
      nextErrors.movementLimitations = movementLimitationsRequiredMessage;
    }
    return nextErrors;
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const validationErrors = validateForm();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    try {
      const saved = await mutation.mutateAsync(payload);
      const next = formStateFromProfile(saved);
      setForm(next);
      setSavedForm(next);
      setErrors({});
      toast.success("Training profile saved");
    } catch (error) {
      setErrors(fieldErrorsFromApiError(error));
    }
  }

  if (profileQuery.isLoading) {
    return (
      <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
        <h1 className="text-2xl font-semibold tracking-tight">
          Training Profile
        </h1>
        <p className="text-sm text-muted-foreground">
          Loading training profile...
        </p>
      </main>
    );
  }

  if (profileQuery.isError) {
    return (
      <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
        <Link
          to="/settings"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          Back to settings
        </Link>
        <h1 className="text-2xl font-semibold tracking-tight">
          Training Profile
        </h1>
        <p className="text-sm text-destructive">
          {getErrorMessage(
            profileQuery.error,
            "Could not load your training profile.",
          )}
        </p>
      </main>
    );
  }

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6">
      <header className="space-y-2">
        <Link
          to="/settings"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          Back to settings
        </Link>
        <h1 className="text-2xl font-semibold tracking-tight">
          Training Profile
        </h1>
        <p className="text-sm text-muted-foreground">{introCopy}</p>
      </header>

      <form
        className="space-y-6"
        onSubmit={handleSubmit}
      >
        <section className="grid gap-4 rounded-lg border bg-card p-4 sm:grid-cols-2">
          <SelectField
            id="primary-goal"
            label="Primary goal"
            value={form.primaryGoal}
            options={goalOptions}
            error={errors.primaryGoal}
            onChange={(primaryGoal) => updateForm({ primaryGoal })}
          />
          <SelectField
            id="experience-level"
            label="Experience level"
            value={form.experienceLevel}
            options={experienceOptions}
            error={errors.experienceLevel}
            onChange={(experienceLevel) => updateForm({ experienceLevel })}
          />
          <FieldShell
            label="Preferred session duration"
            htmlFor="preferred-session-duration"
            error={errors.preferredSessionDurationMinutes}
          >
            <div className="flex items-center gap-3">
              <Input
                id="preferred-session-duration"
                type="number"
                inputMode="numeric"
                min={10}
                max={240}
                value={form.preferredSessionDurationMinutes}
                onChange={(event) =>
                  updateForm({
                    preferredSessionDurationMinutes: event.currentTarget.value,
                  })
                }
                aria-invalid={
                  errors.preferredSessionDurationMinutes ? true : undefined
                }
              />
              <span className="text-sm text-muted-foreground">minutes</span>
            </div>
          </FieldShell>
          <SelectField
            id="usual-training-location"
            label="Usual training location"
            value={form.usualTrainingLocation}
            options={locationOptions}
            error={errors.usualTrainingLocation}
            onChange={(usualTrainingLocation) =>
              updateForm({ usualTrainingLocation })
            }
          />
        </section>

        <section className="space-y-4 rounded-lg border bg-card p-4">
          <TagInput
            label="Available equipment"
            values={form.availableEquipment}
            onChange={(availableEquipment) =>
              updateForm({ availableEquipment })
            }
          />
          <TagInput
            label="Exercises to avoid"
            values={form.avoidedExercises}
            onChange={(avoidedExercises) => updateForm({ avoidedExercises })}
          />
        </section>

        <section className="space-y-4 rounded-lg border bg-card p-4">
          <fieldset className="space-y-3">
            <legend className="text-sm font-medium">
              Movement limitations
            </legend>
            <div className="grid gap-3">
              <RadioOption
                name="movement-limitations"
                label="Not specified"
                checked={form.movementState === "unspecified"}
                onChange={() => updateForm({ movementState: "unspecified" })}
              />
              <RadioOption
                name="movement-limitations"
                label="No known limitations"
                checked={form.movementState === "none"}
                onChange={() => updateForm({ movementState: "none" })}
              />
              <RadioOption
                name="movement-limitations"
                label="I have limitations:"
                checked={form.movementState === "list"}
                onChange={() => updateForm({ movementState: "list" })}
              />
            </div>
            <p className="text-sm text-muted-foreground">
              {movementHelperText}
            </p>
          </fieldset>

          {form.movementState === "list" ? (
            <TagInput
              label="Movement limitation details"
              values={form.movementLimitations}
              error={errors.movementLimitations}
              onChange={(movementLimitations) =>
                updateForm({ movementLimitations })
              }
            />
          ) : null}
        </section>

        {errors.form ? (
          <p className="text-sm text-destructive">{errors.form}</p>
        ) : null}

        <div className="flex justify-end">
          <Button
            type="submit"
            disabled={!isDirty || mutation.isPending}
          >
            <Save className="size-4" />
            {mutation.isPending ? "Saving..." : "Save"}
          </Button>
        </div>
      </form>
    </main>
  );
}

function SelectField({
  id,
  label,
  value,
  options,
  error,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  options: SelectOption[];
  error?: string;
  onChange: (value: string) => void;
}) {
  return (
    <FieldShell
      label={label}
      htmlFor={id}
      error={error}
    >
      <select
        id={id}
        className="border-border bg-background flex h-9 w-full rounded-md border px-3 py-1 text-sm shadow-xs outline-none transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
        value={value}
        onChange={(event) => onChange(event.currentTarget.value)}
        aria-invalid={error ? true : undefined}
      >
        <option value="">Not set</option>
        {options.map((option) => (
          <option
            key={option.value}
            value={option.value}
          >
            {option.label}
          </option>
        ))}
      </select>
    </FieldShell>
  );
}

function FieldShell({
  label,
  htmlFor,
  error,
  children,
}: {
  label: string;
  htmlFor: string;
  error?: string;
  children: ReactNode;
}) {
  return (
    <div className="space-y-2">
      <label
        className="text-sm font-medium"
        htmlFor={htmlFor}
      >
        {label}
      </label>
      {children}
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
    </div>
  );
}

function TagInput({
  label,
  values,
  error,
  onChange,
}: {
  label: string;
  values: string[];
  error?: string;
  onChange: (values: string[]) => void;
}) {
  const [draft, setDraft] = useState("");
  const inputId = label.toLowerCase().replaceAll(" ", "-");

  function addDraft() {
    const next = cleanTagList([...values, draft]);
    if (JSON.stringify(next) !== JSON.stringify(values)) {
      onChange(next);
    }
    setDraft("");
  }

  function removeValue(value: string) {
    onChange(values.filter((current) => current !== value));
  }

  return (
    <div className="space-y-2">
      <label
        className="text-sm font-medium"
        htmlFor={inputId}
      >
        {label}
      </label>
      <Input
        id={inputId}
        value={draft}
        aria-invalid={error ? true : undefined}
        onChange={(event) => setDraft(event.currentTarget.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            addDraft();
          }
        }}
      />
      {values.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {values.map((value) => (
            <span
              key={value}
              className="inline-flex min-h-8 items-center gap-1 rounded-md border bg-secondary px-2 py-1 text-sm text-secondary-foreground"
            >
              {value}
              <button
                type="button"
                className="rounded-sm p-0.5 text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                aria-label={`Remove ${value}`}
                onClick={() => removeValue(value)}
              >
                <X className="size-3" />
              </button>
            </span>
          ))}
        </div>
      ) : null}
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
    </div>
  );
}

function RadioOption({
  name,
  label,
  checked,
  onChange,
}: {
  name: string;
  label: string;
  checked: boolean;
  onChange: () => void;
}) {
  return (
    <label className="flex items-center gap-3 text-sm">
      <input
        type="radio"
        name={name}
        checked={checked}
        onChange={onChange}
      />
      <span>{label}</span>
    </label>
  );
}

function formStateFromProfile(profile: TrainingProfile): FormState {
  return {
    primaryGoal: profile.primary_goal ?? "",
    experienceLevel: profile.experience_level ?? "",
    preferredSessionDurationMinutes:
      profile.preferred_session_duration_minutes?.toString() ?? "",
    usualTrainingLocation: profile.usual_training_location ?? "",
    availableEquipment: profile.available_equipment ?? [],
    avoidedExercises: profile.avoided_exercises ?? [],
    movementState:
      profile.movement_limitations === null
        ? "unspecified"
        : profile.movement_limitations.length === 0
          ? "none"
          : "list",
    movementLimitations: profile.movement_limitations ?? [],
  };
}

function payloadFromForm(form: FormState): UpdateTrainingProfileRequest {
  return {
    primary_goal: form.primaryGoal || null,
    experience_level: form.experienceLevel || null,
    preferred_session_duration_minutes:
      form.preferredSessionDurationMinutes.trim() === ""
        ? null
        : Number(form.preferredSessionDurationMinutes),
    usual_training_location: form.usualTrainingLocation || null,
    available_equipment: cleanTagList(form.availableEquipment),
    avoided_exercises: cleanTagList(form.avoidedExercises),
    movement_limitations:
      form.movementState === "unspecified"
        ? null
        : form.movementState === "none"
          ? []
          : cleanTagList(form.movementLimitations),
  };
}

function cleanTagList(values: string[]): string[] {
  const cleaned: string[] = [];
  const seen = new Set<string>();
  for (const value of values) {
    const trimmed = value.trim().slice(0, 120);
    if (!trimmed) {
      continue;
    }
    const key = trimmed.toLowerCase();
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    cleaned.push(trimmed);
    if (cleaned.length === 20) {
      break;
    }
  }
  return cleaned;
}

function fieldErrorsFromApiError(error: unknown): FieldErrors {
  const message = getErrorMessage(error, "Could not save training profile.");
  if (!isApiError(error)) {
    return { form: message };
  }
  if (message.startsWith("primary_goal:")) {
    return { primaryGoal: message.replace("primary_goal: ", "") };
  }
  if (message.startsWith("experience_level:")) {
    return { experienceLevel: message.replace("experience_level: ", "") };
  }
  if (message.startsWith("preferred_session_duration_minutes:")) {
    return {
      preferredSessionDurationMinutes: message.replace(
        "preferred_session_duration_minutes: ",
        "",
      ),
    };
  }
  if (message.startsWith("usual_training_location:")) {
    return {
      usualTrainingLocation: message.replace("usual_training_location: ", ""),
    };
  }
  return { form: message };
}

export { emptyProfile, formStateFromProfile, goalOptions, payloadFromForm };
