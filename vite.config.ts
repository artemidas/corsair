import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";
import wails from "@wailsio/runtime/plugins/vite";
import { fileURLToPath, URL } from 'node:url'

export default defineConfig(async () => ({
  plugins: [vue(), tailwindcss(), wails("./src/bindings")],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  clearScreen: false,
  server: {
    port: Number(process.env.WAILS_VITE_PORT) || 1420,
    strictPort: true,
    host: "127.0.0.1",
    watch: {
      ignored: ["**/src-tauri/**", "**/*.go"],
    },
  },
}));
