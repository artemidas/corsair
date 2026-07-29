import { ref } from "vue";

export type Theme = "light" | "dark";

const STORAGE_KEY = "corsair-theme";

function applyToDom(theme: Theme) {
  if (typeof document === "undefined") return;
  document.documentElement.classList.toggle("dark", theme === "dark");
}

function readInitial(): Theme {
  if (typeof document === "undefined") return "dark";
  // The inline script in index.html already applied the correct class
  // before this module is evaluated, so just mirror what the DOM says.
  return document.documentElement.classList.contains("dark") ? "dark" : "light";
}

const theme = ref<Theme>(readInitial());

export function useTheme() {
  function setTheme(next: Theme) {
    theme.value = next;
    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch {
      // localStorage may be unavailable (private mode, restricted context)
    }
    applyToDom(next);
  }

  function toggleTheme() {
    setTheme(theme.value === "dark" ? "light" : "dark");
  }

  return { theme, setTheme, toggleTheme };
}
