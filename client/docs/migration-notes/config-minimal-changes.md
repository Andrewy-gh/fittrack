```ts
export default {
  input: 'path/to/your/openapi-spec.json', // or your OpenAPI spec URL
  output: 'src/generated', // or wherever your current generated code is
  plugins: [
    // Your existing plugins...
    {
      name: '@tanstack/react-query',
      queryOptions: {
        // Use "Options" suffix to match your current naming
        name: '{{name}}Options',
        // Keep default camelCase
        case: 'camelCase',
        // Enable metadata if you want operation info
        meta: false, // or customize as needed
      },
      queryKeys: {
        // Generate query keys for cache invalidation
        enabled: true,
        name: '{{name}}QueryKey',
        // Include tags if you want better cache invalidation
        tags: false, // set to true if you want operation tags in keys
      },
      mutationOptions: {
        // For your future mutations
        enabled: true,
        name: '{{name}}Mutation',
      },
      infiniteQueryOptions: {
        // For pagination if needed
        enabled: true,
        name: '{{name}}InfiniteOptions',
      },
      infiniteQueryKeys: {
        enabled: true,
        name: '{{name}}InfiniteQueryKey',
      },
      // Add comments from OpenAPI spec
      comments: true,
      // Don't export from index to avoid conflicts
      exportFromIndex: false,
      // Custom output file name
      output: '@tanstack/react-query',
    }
  ],
};
```