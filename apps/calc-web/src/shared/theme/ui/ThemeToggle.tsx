import { Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";

type Theme = "light" | "dark";

export function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>(() =>
    document.documentElement.dataset.theme === "dark" ? "dark" : "light",
  );

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    try {
      localStorage.setItem("calculator-theme", theme);
    } catch {
      // Theme switching still works when browser storage is unavailable.
    }
  }, [theme]);

  return (
    <div
      className="inline-flex items-center gap-1 rounded-full border border-base-300 bg-base-100 p-1"
      role="group"
      aria-label="Color theme"
    >
      {(["light", "dark"] as const).map((value) => (
        <button
          key={value}
          type="button"
          className={`btn btn-circle btn-sm ${theme === value ? "btn-primary" : "btn-ghost text-base-content/60"}`}
          aria-label={`${value === "light" ? "Light" : "Dark"} theme`}
          aria-pressed={theme === value}
          onClick={() => setTheme(value)}
        >
          {value === "light" ? (
            <Sun size={17} aria-hidden="true" />
          ) : (
            <Moon size={17} aria-hidden="true" />
          )}
        </button>
      ))}
    </div>
  );
}
