import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { formatDate } from '@/lib/utils';

export function WorkoutNotesCard({
  title,
  note,
  dateLabel,
}: {
  title: string;
  note?: string | null;
  dateLabel?: string;
}) {
  const trimmedNote = note?.trim();

  if (!trimmedNote) {
    return null;
  }

  return (
    <Card className="border-0 shadow-sm backdrop-blur-sm">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg font-semibold">{title}</CardTitle>
        {dateLabel && (
          <p className="text-xs text-muted-foreground">
            {formatDate(dateLabel)}
          </p>
        )}
      </CardHeader>
      <CardContent>
        <p className="text-sm leading-6 text-muted-foreground">{trimmedNote}</p>
      </CardContent>
    </Card>
  );
}
