import { useState } from "react";

import { styled } from "@mui/material/styles";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import Grid from "@mui/material/Grid2";
import Link from "@mui/material/Link";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";

import { SourceModel } from "../../client/models";
import { WebAPIClient } from "../../client/WebAPIClient.ts";
import { formatDateWithStyle, parseISOString } from "../../util/dates";
import LabelControl from "./LabelControl.tsx";

export default function Source(props: {
  source: SourceModel;
  sourceUpdated: (id: string, source: SourceModel) => void;
}) {
  const client = new WebAPIClient();
  const source = props.source;

  const [loading, setLoading] = useState<number>(0);

  const saveLabel = (text: string) => {
    setLoading(1);
    client
      .saveLabel(source.Source, source.Label, text)
      .then(
        (resp) => {
          source.Label = resp.Label;
          props.sourceUpdated(source.ID, source);
        },
        (err) => console.log(err)
      )
      .finally(() => setLoading(0));
  };

  const Label = styled(Paper)(({ theme }) => ({
    ...theme.typography.body2,
    padding: theme.spacing(1),
    textAlign: "end",
    color: theme.palette.text.secondary,
    ...theme.applyStyles("dark", {
      backgroundColor: "#1A2027",
    }),
    ...theme.applyStyles("light", {
      backgroundColor: "#EAF0F7",
    }),
  }));

  const Item = styled(Paper)(({ theme }) => ({
    ...theme.typography.body2,
    padding: theme.spacing(1),
    textAlign: "start",
  }));

  return (
    <Box sx={{ width: "100%", maxWidth: { sm: "100%", md: "1700px" } }}>
      <Typography variant="h4" component="h2" sx={{ mb: 2 }}>
        {source.Source} {source.Label}
      </Typography>
      <Grid container spacing={2}>
        <Grid size={{ xs: 4, md: 4 }}>
          <Label>Source</Label>
        </Grid>
        <Grid size={{ xs: 8, md: 8 }}>
          <Item>{source.Source}</Item>
        </Grid>
        <Grid size={{ xs: 4, md: 4 }}>
          <Label>Label</Label>
        </Grid>
        <Grid size={{ xs: 8, md: 8 }}>
          {!!loading ? (
            <CircularProgress />
          ) : (
            <LabelControl
              value={source.Label}
              onTextEntered={saveLabel}
            ></LabelControl>
          )}
        </Grid>
        <Grid size={{ xs: 4, md: 4 }}>
          <Label>Version</Label>
        </Grid>
        <Grid size={{ xs: 8, md: 8 }}>
          <Item>{source.Version}</Item>
        </Grid>
        <Grid size={{ xs: 4, md: 4 }}>
          <Label>Notification URL</Label>
        </Grid>
        <Grid size={{ xs: 8, md: 8 }}>
          <Item>
            <Link href={source.NotificationURL} target="_blank" rel="noopener">
              {source.NotificationURL}
              <sup>🔗</sup>
            </Link>
          </Item>
        </Grid>
        <Grid size={{ xs: 4, md: 4 }}>
          <Label>Repo last updated</Label>
        </Grid>
        <Grid size={{ xs: 8, md: 8 }}>
          <Item>
            {formatDateWithStyle(
              parseISOString(source.Notifications[0].Created),
              "en-gb",
              "longdatetime"
            )}
          </Item>
        </Grid>
      </Grid>
    </Box>
  );
}
