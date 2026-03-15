import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const rawAllowedHosts = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env?.VITE_ALLOWED_HOSTS || ''
const allowedHosts = rawAllowedHosts
  .split(',')
  .map((v: string) => v.trim())
  .filter(Boolean)

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5174,
    allowedHosts: allowedHosts.length > 0 ? allowedHosts : true,
  }
})
