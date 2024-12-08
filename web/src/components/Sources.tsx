import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid2";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import ButtonGroup from "@mui/material/ButtonGroup";
import Button from "@mui/material/Button";
import RpcClientService from "../client/rpcClientService.ts";
import { Source } from "../client/models.ts";
import { useEffect, useState } from "react";
import Checkbox from "@mui/material/Checkbox";

export default function Sources() {
  const [rows, setRows] = useState<Source[]>([]);
  const [selectedIDs, setSelectedIDs] = useState<string[]>([]);
  const [refresh, setRefresh] = useState<number>(0);

  useEffect(() => {
    const rpcService = new RpcClientService();
    rpcService.execute<Source[]>("GetSources").then(
      (ss) => {
        if (!ss) {
          setRows([]);
          return;
        }
        setRows(ss);
      },
      () => setRows([])
    );
  }, []);

  const handleClick = (row: Source) => {
    const key = row.Source + "." + row.Label;
    const idx = selectedIDs.indexOf(key);
    if (idx < 0) {
      selectedIDs.push(key);
    } else {
      selectedIDs.splice(idx, 1);
    }
    setSelectedIDs(selectedIDs);
    setRefresh(refresh ^ 1);
  };

  return (
    <Box sx={{ width: "100%", maxWidth: { sm: "100%", md: "1700px" } }}>
      <Typography variant="h4" component="h1" sx={{ mb: 2 }}>
        Sources
      </Typography>
      <ButtonGroup variant="outlined" aria-label="Basic button group">
        <Button>Label</Button>
        <Button>Update</Button>
      </ButtonGroup>
      <Grid container spacing={2} columns={12}>
        <Grid size={{ xs: 12, lg: 12 }}>
          <TableContainer component={Paper}>
            <Table
              sx={{ minWidth: 700 }}
              aria-label="customized table"
              size={"medium"}
            >
              <TableHead>
                <TableRow>
                  <TableCell padding="checkbox"></TableCell>
                  <TableCell component="th" scope="row" padding="normal">
                    Source
                  </TableCell>
                  <TableCell component="th" scope="row" padding="normal">
                    Label
                  </TableCell>
                  <TableCell
                    align="right"
                    component="th"
                    scope="row"
                    padding="normal"
                  >
                    Version
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {rows.map((row, index) => {
                  const isItemSelected = selectedIDs.includes(
                    row.Source + "." + row.Label
                  );
                  const labelId = `enhanced-table-checkbox-${index}`;
                  return (
                    <TableRow
                      hover
                      onClick={() => handleClick(row)}
                      role="checkbox"
                      aria-checked={isItemSelected}
                      tabIndex={-1}
                      key={refresh}
                      selected={isItemSelected}
                      sx={{ cursor: "pointer" }}
                    >
                      <TableCell padding="checkbox">
                        <Checkbox
                          color="primary"
                          checked={isItemSelected}
                          inputProps={{
                            "aria-labelledby": labelId,
                          }}
                        />
                      </TableCell>
                      <TableCell
                        component="th"
                        id={labelId}
                        scope="row"
                        padding="normal"
                      >
                        {row.Source}
                      </TableCell>
                      <TableCell component="th" scope="row" padding="normal">
                        {row.Label}
                      </TableCell>
                      <TableCell
                        align="right"
                        component="th"
                        scope="row"
                        padding="normal"
                      >
                        {row.Version}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>
        </Grid>
      </Grid>
    </Box>
  );
}
