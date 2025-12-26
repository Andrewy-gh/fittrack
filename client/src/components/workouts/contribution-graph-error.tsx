import { AlertCircle } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';

export function ContributionGraphError() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Contribution Graph</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-3 text-muted-foreground">
          <AlertCircle className="w-5 h-5 flex-shrink-0" />
          <p className="text-sm">
            Unable to load contribution graph. Please try refreshing the page.
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
