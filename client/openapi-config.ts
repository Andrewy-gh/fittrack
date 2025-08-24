/**
 * Configuration for the OpenAPI TypeScript Codegen
 * @see https://github.com/ferdikoomen/openapi-typescript-codegen
 */

const config = {
  // Input and output paths
  input: '../server/docs/swagger.json',
  output: 'src/generated',
  
  // Client configuration
  client: 'axios',
  
  // Naming conventions
  useOptions: true,
  useUnionTypes: true,
  
  // Type generation options
  exportCore: true,
  exportServices: true,
  exportModels: true,
  exportSchemas: false,
  
  // Code style options
  indent: '  ', // 2 spaces
  postfixServices: 'Service',
  postfixModels: '',
  
  // Request/Response configuration
  request: './src/lib/api-client.ts', // Optional: custom axios instance
  
  // Additional options
  write: true,
  format: true,
  lint: false,
};

export default config;
