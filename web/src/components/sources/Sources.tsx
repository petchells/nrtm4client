import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid2";
import RpcClientService from "../../client/rpcClientService.ts";
import { SourceModel } from "../../client/models.ts";
import { useEffect, useState } from "react";
import Alert from "@mui/material/Alert";
import WarningIcon from "@mui/icons-material/Warning";
import SourcesTable from "./SourcesTable.tsx";

export default function Sources() {
  const [err, setErr] = useState<string>("");
  const [rows, setRows] = useState<SourceModel[]>([]);
  const [selectedIDs, setSelectedIDs] = useState<string[]>([]);

  useEffect(() => {
    const rpcService = new RpcClientService();
    rpcService.execute<SourceModel[]>("GetSources").then(
      (ss) => {
        if (!ss) {
          setRows([]);
          return;
        }
        setRows(ss);
      },
      () => {
        setErr("No connection to the server");
        setRows([]);
      }
    );
  }, []);

  const handleOnSelected = (row: SourceModel) => {
    const key = row.Source + "." + row.Label;
    const idx = selectedIDs.indexOf(key);
    if (idx < 0) {
      selectedIDs.push(key);
    } else {
      selectedIDs.splice(idx, 1);
    }
    setSelectedIDs(selectedIDs);
  };

  const errorContainer = (err: string) => {
    return (
      <>
        <Alert icon={<WarningIcon fontSize="inherit" />} severity="warning">
          Cannot connect to backend: {err}
        </Alert>
      </>
    );
  };

  const noSources = () => {
    return (
      <>
        <Alert icon={<WarningIcon fontSize="inherit" />} severity="success">
          No sources
        </Alert>
        <Typography variant="h4" component="p" sx={{ mb: 2 }}>
          Sources
        </Typography>
      </>
    );
  };

  return (
    <Box sx={{ width: "100%", maxWidth: { sm: "100%", md: "1700px" } }}>
      <Typography variant="h4" component="h1" sx={{ mb: 2 }}>
        Sources
      </Typography>
      <Grid container spacing={2} columns={12}>
        {err
          ? errorContainer(err)
          : rows.length
          ? SourcesTable({
              rows: rows,
              selectedIDs: selectedIDs,
              onSelected: (row: SourceModel) => handleOnSelected(row),
            })
          : noSources()}
        {/* <ButtonGroup variant="outlined" aria-label="Actions for source">
        <Button>Label</Button>
        <Button>Update</Button>
      </ButtonGroup>
 */}
      </Grid>
    </Box>
  );
}
