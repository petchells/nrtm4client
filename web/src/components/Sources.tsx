import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import Stack from "@mui/material/Stack";
import RpcClientService from "../client/rpcClientService.ts";
import { Source } from "../client/models.ts";
import { useState } from "react";

export default function Sources() {
  const rpcService = new RpcClientService();
  const [sources, setSources] = useState([] as Source[]);
  rpcService.execute<Source[]>("GetSources").then(
    (ss) => {
      if (ss) {
        setSources(ss);
      }
    },
    () => setSources([])
  );

  return (
    <Box sx={{ width: "100%", maxWidth: { sm: "100%", md: "1700px" } }}>
      <Typography variant="h4" component="h1" sx={{ mb: 2 }}>
        Sources
      </Typography>
      <Stack sx={{ flexGrow: 1, p: 1, justifyContent: "space-between" }}>
        <List dense>
          {sources.map((src, index) => (
            <ListItem key={index} disablePadding sx={{ display: "block" }}>
              <ListItemText primary={src.Source} />
            </ListItem>
          ))}
        </List>
      </Stack>
    </Box>
  );
}
