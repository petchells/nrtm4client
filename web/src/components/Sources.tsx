import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid2";
import { DataGrid } from "@mui/x-data-grid";
import ButtonGroup from "@mui/material/ButtonGroup";
import Button from "@mui/material/Button";
import { GridColDef } from "@mui/x-data-grid";
import RpcClientService from "../client/rpcClientService.ts";
import { Source } from "../client/models.ts";
import { useEffect, useState } from "react";

const columns: GridColDef[] = [
  {
    field: "Source",
    headerName: "Source",
    flex: 1,
    minWidth: 60,
  },
  {
    field: "Label",
    headerName: "Label",
    flex: 1,
    minWidth: 80,
  },
  {
    field: "Version",
    headerName: "Version",
    headerAlign: "right",
    align: "right",
    flex: 1,
    minWidth: 20,
  },
  {
    field: "NotificationURL",
    headerName: "URL",
    flex: 2,
    minWidth: 200,
  },
];

export default function Sources() {
  const [rows, setRows] = useState([] as Source[]);
  const [selectedIDs, setSelectedIDs] = useState([] as number[]);

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
          <DataGrid
            checkboxSelection
            rows={rows}
            columns={columns}
            getRowId={(row) => row.ID}
            getRowClassName={(params) =>
              params.indexRelativeToCurrentPage % 2 === 0 ? "even" : "odd"
            }
            density="compact"
            initialState={{
              pagination: { paginationModel: { pageSize: 20 } },
            }}
            pageSizeOptions={[10, 20, 50]}
            slotProps={{
              filterPanel: {
                filterFormProps: {
                  logicOperatorInputProps: {
                    variant: "outlined",
                    size: "small",
                  },
                  columnInputProps: {
                    variant: "outlined",
                    size: "small",
                    sx: { mt: "auto" },
                  },
                  operatorInputProps: {
                    variant: "outlined",
                    size: "small",
                    sx: { mt: "auto" },
                  },
                  valueInputProps: {
                    InputComponentProps: {
                      variant: "outlined",
                      size: "small",
                    },
                  },
                },
              },
            }}
          />
        </Grid>
      </Grid>
    </Box>
  );
}
