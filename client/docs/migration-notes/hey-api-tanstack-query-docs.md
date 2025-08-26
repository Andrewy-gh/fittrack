Build better APIs with Hey API Platform [Dashboard](https://app.heyapi.dev/)
[Skip to content](https://heyapi.dev/openapi-ts/plugins/tanstack-query#VPContent)
[Hey API](https://heyapi.dev/)
Search
Appearance
Menu
Table of Contents
Are you an LLM? You can read better optimized documentation at /openapi-ts/plugins/tanstack-query.md for this page in Markdown format
# TanStack Query v5
v5
### About [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#about)
is a powerful asynchronous state management solution for TypeScript/JavaScript, React, Solid, Vue, Svelte, and Angular.
The TanStack Query plugin for Hey API generates functions and query keys from your OpenAPI spec, fully compatible with SDKs, transformers, and all core features.
### Demo [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#demo)
Launch demo 
## Features [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#features)
  * TanStack Query v5 support
  * seamless integration with `@hey-api/openapi-ts` ecosystem
  * create query keys following the best practices
  * type-safe query options, infinite query options, and mutation options
  * minimal learning curve thanks to extending the underlying technology


## Installation [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#installation)
In your [configuration](https://heyapi.dev/openapi-ts/get-started), add TanStack Query to your plugins and you'll be ready to generate TanStack Query artifacts. 🎉
reactvueangularsveltesolid
js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  '@tanstack/react-query', 
 ],
};
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  '@tanstack/vue-query', 
 ],
};
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  '@tanstack/angular-query-experimental', 
 ],
};
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  '@tanstack/svelte-query', 
 ],
};
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  '@tanstack/solid-query', 
 ],
};
```

## Output [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#output)
The TanStack Query plugin will generate the following artifacts, depending on the input specification.
## Queries [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#queries)
Queries are generated from GET and POST endpoints. The generated query functions follow the naming convention of SDK functions and by default append `Options`, e.g. `getPetByIdOptions()`.
exampleconfig
ts```
const { data, error } = useQuery({
 ...getPetByIdOptions({
  path: {
   petId: 1,
  },
 }),
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   queryOptions: true, 
  },
 ],
};
```

You can customize the naming and casing pattern for `queryOptions` functions using the `.name` and `.case` options.
### Meta [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#meta)
You can use the `meta` field to attach arbitrary information to a query. To generate metadata for `queryOptions`, provide a function to the `.meta` option.
exampleconfig
ts```
queryOptions({
 // ...other fields
 meta: {
  id: 'getPetById',
 },
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   queryOptions: {
    meta: (operation) => ({ id: operation.id }), 
   },
  },
 ],
};
```

## Query Keys [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#query-keys)
Query keys contain normalized SDK function parameters and additional metadata.
exampleconfig
ts```
const queryKey = [
 {
  _id: 'getPetById',
  baseUrl: 'https://app.heyapi.dev',
  path: {
   petId: 1,
  },
 },
];
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   queryKeys: true, 
  },
 ],
};
```

### Tags [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#tags)
You can include operation tags in your query keys by setting `tags` to `true`. This will make query keys larger but provides better cache invalidation capabilities.
exampleconfig
ts```
const queryKey = [
 {
  _id: 'getPetById',
  baseUrl: 'https://app.heyapi.dev',
  path: {
   petId: 1,
  },
  tags: ['pets', 'one', 'get'], 
 },
];
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   queryKeys: {
    tags: true, 
   },
  },
 ],
};
```

### Accessing Query Keys [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#accessing-query-keys)
If you have access to the result of query options function, you can get the query key from the `queryKey` field.
exampleconfig
ts```
const { queryKey } = getPetByIdOptions({
 path: {
  petId: 1,
 },
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   queryOptions: true, 
  },
 ],
};
```

Alternatively, you can access the same query key by calling query key functions. The generated query key functions follow the naming convention of SDK functions and by default append `QueryKey`, e.g. `getPetByIdQueryKey()`.
exampleconfig
ts```
const queryKey = getPetByIdQueryKey({
 path: {
  petId: 1,
 },
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   queryKeys: true, 
  },
 ],
};
```

You can customize the naming and casing pattern for `queryKeys` functions using the `.name` and `.case` options.
## Infinite Queries [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#infinite-queries)
Infinite queries are generated from GET and POST endpoints if we detect a [pagination](https://heyapi.dev/openapi-ts/configuration/parser#pagination) parameter. The generated infinite query functions follow the naming convention of SDK functions and by default append `InfiniteOptions`, e.g. `getFooInfiniteOptions()`.
exampleconfig
ts```
const { data, error } = useInfiniteQuery({
 ...getFooInfiniteOptions({
  path: {
   fooId: 1,
  },
 }),
 getNextPageParam: (lastPage, pages) => lastPage.nextCursor,
 initialPageParam: 0,
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   infiniteQueryOptions: true, 
  },
 ],
};
```

You can customize the naming and casing pattern for `infiniteQueryOptions` functions using the `.name` and `.case` options.
### Meta [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#meta-1)
You can use the `meta` field to attach arbitrary information to a query. To generate metadata for `infiniteQueryOptions`, provide a function to the `.meta` option.
exampleconfig
ts```
infiniteQueryOptions({
 // ...other fields
 meta: {
  id: 'getPetById',
 },
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   infiniteQueryOptions: {
    meta: (operation) => ({ id: operation.id }), 
   },
  },
 ],
};
```

## Infinite Query Keys [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#infinite-query-keys)
Infinite query keys contain normalized SDK function parameters and additional metadata.
exampleconfig
ts```
const queryKey = [
 {
  _id: 'getPetById',
  _infinite: true,
  baseUrl: 'https://app.heyapi.dev',
  path: {
   petId: 1,
  },
 },
];
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   infiniteQueryKeys: true, 
  },
 ],
};
```

### Tags [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#tags-1)
You can include operation tags in your infinite query keys by setting `tags` to `true`. This will make query keys larger but provides better cache invalidation capabilities.
exampleconfig
ts```
const queryKey = [
 {
  _id: 'getPetById',
  _infinite: true,
  baseUrl: 'https://app.heyapi.dev',
  path: {
   petId: 1,
  },
  tags: ['pets', 'one', 'get'], 
 },
];
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   infiniteQueryKeys: {
    tags: true, 
   },
  },
 ],
};
```

### Accessing Infinite Query Keys [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#accessing-infinite-query-keys)
If you have access to the result of infinite query options function, you can get the query key from the `queryKey` field.
exampleconfig
ts```
const { queryKey } = getPetByIdInfiniteOptions({
 path: {
  petId: 1,
 },
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   infiniteQueryOptions: true, 
  },
 ],
};
```

Alternatively, you can access the same query key by calling query key functions. The generated query key functions follow the naming convention of SDK functions and by default append `InfiniteQueryKey`, e.g. `getPetByIdInfiniteQueryKey()`.
exampleconfig
ts```
const queryKey = getPetByIdInfiniteQueryKey({
 path: {
  petId: 1,
 },
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   infiniteQueryKeys: true, 
  },
 ],
};
```

You can customize the naming and casing pattern for `infiniteQueryKeys` functions using the `.name` and `.case` options.
## Mutations [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#mutations)
Mutations are generated from DELETE, PATCH, POST, and PUT endpoints. The generated mutation functions follow the naming convention of SDK functions and by default append `Mutation`, e.g. `addPetMutation()`.
exampleconfig
ts```
const addPet = useMutation({
 ...addPetMutation(),
 onError: (error) => {
  console.log(error);
 },
});
addPet.mutate({
 body: {
  name: 'Kitty',
 },
});
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   mutationOptions: true, 
  },
 ],
};
```

You can customize the naming and casing pattern for `mutationOptions` functions using the `.name` and `.case` options.
### Meta [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#meta-2)
You can use the `meta` field to attach arbitrary information to a mutation. To generate metadata for `mutationOptions`, provide a function to the `.meta` option.
exampleconfig
ts```
const mutationOptions = {
 // ...other fields
 meta: {
  id: 'getPetById',
 },
};
```

js```
export default {
 input: 'hey-api/backend', // sign up at app.heyapi.dev
 output: 'src/client',
 plugins: [
  // ...other plugins
  {
   name: '@tanstack/react-query',
   mutationOptions: {
    meta: (operation) => ({ id: operation.id }), 
   },
  },
 ],
};
```

## API [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#api)
You can view the complete list of options in the interface.
## Examples [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#examples)
You can view live examples on .
## Sponsors [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#sponsors)
Help Hey API stay around for the long haul by becoming a .
### Gold [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#gold)
  * Best-in-class SDKs and MCP for your API. 


### Silver [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#silver)
### Bronze [​](https://heyapi.dev/openapi-ts/plugins/tanstack-query#bronze)
