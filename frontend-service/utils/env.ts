// Environment variable validation
const requiredEnvVars = [
  "NEXT_PUBLIC_SUPABASE_URL",
  "NEXT_PUBLIC_SUPABASE_ANON_KEY",
  "NEXT_PUBLIC_BACKEND_URL",
] as const;

// Export environment variables with type safety and defaults
export const env = {
  backendUrl: process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080",
  supabaseUrl: process.env.NEXT_PUBLIC_SUPABASE_URL || "",
  supabaseAnonKey: process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || "",
  vercelEnv: process.env.VERCEL_ENV || "development",
  vercelUrl: process.env.VERCEL_URL || "localhost:3000",
  vercelProjectUrl:
    process.env.VERCEL_PROJECT_PRODUCTION_URL || "your-project.vercel.app",
} as const;

// Export a function to check if all required environment variables are present
export function hasRequiredEnvVars(): boolean {
  return requiredEnvVars.every((envVar) => process.env[envVar]);
}

// Export a function to get missing environment variables
export function getMissingEnvVars(): string[] {
  return requiredEnvVars.filter((envVar) => !process.env[envVar]);
}
