import { createFileRoute } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import type { AnyFieldApi } from '@tanstack/react-form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export const Route = createFileRoute('/form')({
  component: RouteComponent,
})

export function FieldInfo({ field }: { field: AnyFieldApi }) {
  return (
    <>
      {field.state.meta.isTouched && !field.state.meta.isValid ? (
        <p className="text-sm text-destructive mt-1">{field.state.meta.errors.join(', ')}</p>
      ) : null}
      {field.state.meta.isValidating ? (
        <p className="text-sm text-muted-foreground mt-1">Validating...</p>
      ) : null}
    </>
  )
}

export function SimpleFormExample() {
  const form = useForm({
    defaultValues: {
      firstName: '',
      lastName: '',
    },
    onSubmit: async ({ value }) => {
      // Do something with form data
      console.log(value)
    },
  })

  return (
    <div className="container mx-auto py-10">
      <Card className="max-w-md mx-auto">
        <CardHeader>
          <CardTitle>Simple Form Example</CardTitle>
          <CardDescription>
            A form built with TanStack Form and shadcn/ui components
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
            className="space-y-4"
          >
            <div className="space-y-2">
              <form.Field
                name="firstName"
                validators={{
                  onChange: ({ value }) =>
                    !value
                      ? 'A first name is required'
                      : value.length < 3
                        ? 'First name must be at least 3 characters'
                        : undefined,
                  onChangeAsyncDebounceMs: 500,
                  onChangeAsync: async ({ value }) => {
                    await new Promise((resolve) => setTimeout(resolve, 1000))
                    return (
                      value.includes('error') && 'No "error" allowed in first name'
                    )
                  },
                }}
                children={(field) => (
                  <>
                    <Label htmlFor={field.name}>First Name</Label>
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      className={field.state.meta.isTouched && !field.state.meta.isValid ? 'border-destructive' : ''}
                    />
                    <FieldInfo field={field} />
                  </>
                )}
              />
            </div>
            <div className="space-y-2">
              <form.Field
                name="lastName"
                children={(field) => (
                  <>
                    <Label htmlFor={field.name}>Last Name</Label>
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      className={field.state.meta.isTouched && !field.state.meta.isValid ? 'border-destructive' : ''}
                    />
                    <FieldInfo field={field} />
                  </>
                )}
              />
            </div>
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit} className="w-full">
                  {isSubmitting ? 'Submitting...' : 'Submit'}
                </Button>
              )}
            />
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

function RouteComponent() {
  return <SimpleFormExample />
}
