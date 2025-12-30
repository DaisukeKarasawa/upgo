import typography from "@tailwindcss/typography";

/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      typography: {
        DEFAULT: {
          css: {
            fontWeight: "300",
            "--tw-prose-body": "rgb(55, 65, 81)",
            "--tw-prose-headings": "rgb(17, 24, 39)",
            "--tw-prose-links": "rgb(55, 65, 81)",
            "--tw-prose-bold": "rgb(17, 24, 39)",
            "--tw-prose-code": "rgb(31, 41, 55)",
            "--tw-prose-pre-code": "rgb(31, 41, 55)",
            "--tw-prose-pre-bg": "rgb(249, 250, 251)",
            "--tw-prose-th-borders": "rgb(209, 213, 219)",
            "--tw-prose-td-borders": "rgb(209, 213, 219)",
          },
        },
      },
      keyframes: {
        drawLineTop: {
          from: { transform: "scaleX(0)" },
          to: { transform: "scaleX(1)" },
        },
        drawLineRight: {
          from: { transform: "scaleY(0)" },
          to: { transform: "scaleY(1)" },
        },
        drawLineBottom: {
          from: { transform: "scaleX(0)" },
          to: { transform: "scaleX(1)" },
        },
        drawLineLeft: {
          from: { transform: "scaleY(0)" },
          to: { transform: "scaleY(1)" },
        },
      },
      animation: {
        drawLineTop: "drawLineTop 0.15s ease-out forwards",
        drawLineRight: "drawLineRight 0.15s ease-out forwards",
        drawLineBottom: "drawLineBottom 0.15s ease-out forwards",
        drawLineLeft: "drawLineLeft 0.15s ease-out forwards",
      },
    },
  },
  plugins: [typography],
};
