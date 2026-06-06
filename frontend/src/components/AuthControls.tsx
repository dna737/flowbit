import { Show, SignInButton, SignUpButton, UserButton } from "@clerk/react";
import { Button, Stack } from "@mui/material";

import { tokens } from "../theme";

export function AuthControls() {
  return (
    <Stack direction="row" spacing={1} alignItems="center">
      <Show when="signed-out">
        <SignInButton mode="modal">
          <Button
            size="small"
            variant="outlined"
            sx={{
              height: 32,
              borderColor: tokens.color.borderDefault,
              color: tokens.color.textPrimary,
            }}
          >
            Sign in
          </Button>
        </SignInButton>
        <SignUpButton mode="modal">
          <Button
            size="small"
            variant="contained"
            sx={{
              height: 32,
              backgroundColor: tokens.color.accentBlue,
              color: "#fff",
              "&:hover": { backgroundColor: "#2563EB" },
            }}
          >
            Sign up
          </Button>
        </SignUpButton>
      </Show>
      <Show when="signed-in">
        <UserButton />
      </Show>
    </Stack>
  );
}
