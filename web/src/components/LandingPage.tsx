import { RouterProvider } from "react-router-dom";
import type {} from "@mui/x-date-pickers/themeAugmentation";
import type {} from "@mui/x-charts/themeAugmentation";
import type {} from "@mui/x-data-grid/themeAugmentation";
import type {} from "@mui/x-tree-view/themeAugmentation";
import { alpha } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import AppNavbar from "./widgets/AppNavbar";
import Header from "./widgets/Header";
import SideMenu from "./widgets/SideMenu";
import AppTheme from "./shared-theme/AppTheme";
import AnalyticsRoundedIcon from "@mui/icons-material/AnalyticsRounded";
import AssignmentRoundedIcon from "@mui/icons-material/AssignmentRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import InfoRoundedIcon from "@mui/icons-material/InfoRounded";
import HelpRoundedIcon from "@mui/icons-material/HelpRounded";
import {
  chartsCustomizations,
  dataGridCustomizations,
  datePickersCustomizations,
  treeViewCustomizations,
} from "../theme/customizations";

import { useState } from "react";
import { FileDownload, QuestionAnswer } from "@mui/icons-material";
import { router } from "./root";
import {} from "react-router-dom";

const xThemeComponents = {
  ...chartsCustomizations,
  ...dataGridCustomizations,
  ...datePickersCustomizations,
  ...treeViewCustomizations,
};
const mainListItems = [
  { text: "Sources", icon: <FileDownload />, href: "/" },
  { text: "Object queries", icon: <QuestionAnswer />, href: "/queries" },
  { text: "Dashboard", icon: <AnalyticsRoundedIcon />, href: "/dashboard" },
  { text: "Tasks", icon: <AssignmentRoundedIcon /> },
];

const secondaryListItems = [
  { text: "Settings", icon: <SettingsRoundedIcon /> },
  { text: "About", icon: <InfoRoundedIcon /> },
  { text: "Feedback", icon: <HelpRoundedIcon /> },
];

export default function LandingPage(props: { disableCustomTheme?: boolean }) {
  const [menuItemSelectedIdx, setMenuItemSelectedIdx] = useState(0);
  return (
    <AppTheme {...props} themeComponents={xThemeComponents}>
      <CssBaseline enableColorScheme />
      <Box sx={{ display: "flex" }}>
        <SideMenu
          mainItems={mainListItems}
          secondaryItems={secondaryListItems}
          onSelected={(idx) => {
            setMenuItemSelectedIdx(idx);
          }}
          menuItemSelectedIdx={menuItemSelectedIdx}
        />
        <AppNavbar pageTitle="Browser" />
        {/* Main content */}
        <Box
          component="main"
          sx={(theme) => ({
            flexGrow: 1,
            backgroundColor: theme.vars
              ? `rgba(${theme.vars.palette.background.defaultChannel} / 1)`
              : alpha(theme.palette.background.default, 1),
            overflow: "auto",
          })}
        >
          <Stack
            spacing={2}
            sx={{
              alignItems: "center",
              mx: 3,
              pb: 5,
              mt: { xs: 8, md: 0 },
            }}
          >
            <Header />
            <RouterProvider router={router} />
          </Stack>
        </Box>
      </Box>
    </AppTheme>
  );
}
