import { createTheme } from "@mui/material/styles";

export const theme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: "#1256d8",
    },
    secondary: {
      main: "#0f9d7a",
    },
    background: {
      default: "#f5f7fb",
      paper: "#ffffff",
    },
  },
  shape: {
    borderRadius: 14,
  },
  typography: {
    fontFamily: '"Segoe UI", "Helvetica Neue", Arial, sans-serif',
    h4: {
      fontWeight: 700,
    },
    h6: {
      fontWeight: 700,
    },
  },
});
