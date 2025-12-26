# ErrorComponent component

The ErrorComponent component is a component that renders an error message and optionally the error's message.
## [ErrorComponent props](https://tanstack.com/router/v1/docs/framework/react/api/router/errorComponentComponent#errorcomponent-props)
The ErrorComponent component accepts the following props:
### [props.error prop](https://tanstack.com/router/v1/docs/framework/react/api/router/errorComponentComponent#propserror-prop)
  * Type: TError (defaults to Error)
  * The error that was thrown by the component's children

### [props.info prop](https://tanstack.com/router/v1/docs/framework/react/api/router/errorComponentComponent#propsinfo-prop)
  * Type: { componentStack: string }
  * Optional
  * Additional information about where the error was thrown, such as the React component stack trace.

### [props.reset prop](https://tanstack.com/router/v1/docs/framework/react/api/router/errorComponentComponent#propsreset-prop)
  * Type: () => void
  * A function to programmatically reset the error state

## [ErrorComponent returns](https://tanstack.com/router/v1/docs/framework/react/api/router/errorComponentComponent#errorcomponent-returns)
  * Returns a formatted error message with the error's message if it exists.
  * The error message can be toggled by clicking the "Show Error" button.
  * By default, the error message will be shown in development.