/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
        spacing: {
            '128': '32rem',
            '144': '36rem',
            '160': '40rem',
            '192': '48rem',
        },
        rotate: {
          '135': '135deg',
        },
        animation: {
            "scale-in-hor-center": "scale-in-hor-center 0.5s cubic-bezier(0.250, 0.460, 0.450, 0.940)   both",
            "scale-out-horizontal": "scale-out-horizontal 0.5s cubic-bezier(0.550, 0.085, 0.680, 0.530)   both",
            "pulsate-fwd": "pulsate-fwd 0.3s ease infinite both",
        },
        keyframes: {
            "scale-in-hor-center": {
                "0%": {
                    transform: "scaleX(0)",
                    opacity: "1"
                },
                to: {
                    transform: "scaleX(1)",
                    opacity: "1"
                }
            },
            "scale-out-horizontal": {
              "0%": {
                  transform: "scaleX(1)",
                  opacity: "1"
              },
              to: {
                  transform: "scaleX(0)",
                  opacity: "1"
              }
            },
            "pulsate-fwd": {
                "0%": {
                    transform: "scale(1)",
                },
                to: {
                    transform: "scale(1.05)",
                }
            }
        }
    }
  },
  plugins: [],
}

