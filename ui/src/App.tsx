import React, { useEffect, useRef, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardActions,
  CardContent,
  Container,
  FormControl,
  FormControlLabel,
  FormGroup,
  FormLabel,
  Radio,
  RadioGroup,
  Slider,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import settings from "./settings";
import styles from "./styles";

interface Job {
  id: string;
  name: string;
  status?: string;
  message?: string;
}

const WS_HOST =
  process.env.NODE_ENV === "development"
    ? "localhost:8090"
    : document.location.host;

function App() {
  const [formData, setFormData] = useState({ name: "", ...settings.defaults });
  const [error, setError] = useState("");
  const [scanning, setScanning] = useState(false);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [connect, setConnect] = useState(0);

  const ws = useRef<null | WebSocket>(null);
  const sendQueue = useRef<string[]>([]);
  const jobsData = useRef<Record<string, Job>>({});

  useEffect(() => {
    const intervalId = setInterval(blitJobs, 250);
    return () => clearInterval(intervalId);
  }, []);

  useEffect(() => {
    ws.current = new WebSocket(`ws://${WS_HOST}/ws`);
    ws.current.onopen = () => sendRequests.current();
    ws.current.onmessage = (event) => {
      const { id, name, status, message } = JSON.parse(event.data);
      if (id) {
        jobsData.current[id] = { id, name, status, message };
        if (message.includes("scanning done")) {
          setScanning(false);
        }
      } else if (status === "failed") {
        setError(message);
        setScanning(false);
      }
    };

    const wsCurrent = ws.current;

    return () => {
      if (wsCurrent != null) {
        wsCurrent.close();
      }
    };
  }, [connect]);

  // Sends queued requests to the websocket. This is a ref so that it can be called from an effect.
  const sendRequests = useRef(() => {
    if (ws.current == null) {
      return;
    }

    if (
      ws.current.readyState === WebSocket.CLOSING ||
      ws.current.readyState === WebSocket.CLOSED
    ) {
      ws.current = null;
      setConnect(connect + 1);
      return;
    }

    let request = sendQueue.current.shift();

    while (request) {
      ws.current.send(request);
      request = sendQueue.current.shift();
    }
  });

  const sendRequest = (request: string) => {
    sendQueue.current.push(request);
    sendRequests.current();
  };

  const blitJobs = () => {
    setJobs(Object.values(jobsData.current).reverse());
  };

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const input = event.target as HTMLInputElement;
    const value = input.type === "checkbox" ? input.checked : input.value;
    setFormData({ ...formData, [input.name]: value });
  };

  const handleScan = () => {
    setError("");
    setScanning(true);
    const { name, ...settings } = formData;
    sendRequest(JSON.stringify({ name, settings }));
  };

  const handleDismiss = (jobId: string) => {
    delete jobsData.current[jobId];
    blitJobs();
  };

  return (
    <Container maxWidth="md">
      <Stack
        direction="row"
        sx={{
          flexWrap: "wrap",
          justifyContent: "space-evenly",
          "> *": { marginTop: "16px" },
          "&:last-child": { marginBottom: "16px" },
        }}
      >
        <FormGroup>
          <Stack spacing={2}>
            <FormControl>
              <FormLabel id="source-label">Page</FormLabel>
              <RadioGroup
                aria-labelledby="source-label"
                value={formData.source}
                name="source"
                onChange={handleChange}
                row
              >
                {settings.available.sources.map(({ value, label }) => (
                  <FormControlLabel
                    key={value}
                    value={value}
                    label={label}
                    control={<Radio />}
                  />
                ))}
              </RadioGroup>
            </FormControl>

            <FormControl>
              <FormLabel id="mode-label">Mode</FormLabel>
              <RadioGroup
                aria-labelledby="mode-label"
                value={formData.mode}
                name="mode"
                onChange={handleChange}
                row
              >
                {settings.available.modes.map(({ value, label }) => (
                  <FormControlLabel
                    key={value}
                    value={value}
                    label={label}
                    control={<Radio />}
                  />
                ))}
              </RadioGroup>
            </FormControl>

            <FormControl>
              <FormLabel id="resolution-label">Resolution</FormLabel>
              <Box sx={{ margin: "0 10px" }}>
                <Slider
                  aria-labelledby="resolution-label"
                  min={settings.available.resolutions[0].value}
                  max={
                    settings.available.resolutions[
                      settings.available.resolutions.length - 1
                    ].value
                  }
                  value={formData.resolution}
                  onChange={(event: Event, value: number | number[]) =>
                    setFormData({
                      ...formData,
                      resolution: value as number,
                    })
                  }
                  step={null}
                  marks={settings.available.resolutions}
                />
              </Box>
            </FormControl>

            <FormGroup row>
              <FormControlLabel
                control={
                  <Switch
                    name={"clean"}
                    checked={formData.clean}
                    onChange={handleChange}
                  />
                }
                label="Auto Clean"
              />

              <FormControlLabel
                control={
                  <Switch
                    name={"pdf"}
                    checked={formData.pdf}
                    onChange={handleChange}
                  />
                }
                label="PDF"
              />
            </FormGroup>

            <TextField
              id="file-name"
              name="name"
              label="File Name"
              variant="outlined"
              value={formData.name}
              onChange={handleChange}
            />

            <Button
              variant="contained"
              disabled={scanning}
              onClick={handleScan}
            >
              {scanning ? "Scanning..." : "Scan"}
            </Button>

            {error && (
              <Alert onClose={() => setError("")} severity="error">
                {error}
              </Alert>
            )}
          </Stack>
        </FormGroup>

        <Stack sx={{ width: "350px" }} spacing={2}>
          {jobs.map((job) => (
            <Card key={job.id}>
              <CardContent sx={{ "&:last-child": { paddingBottom: "8px" } }}>
                <Typography variant="body2">{job.name}</Typography>
                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={job.status === "in_progress" ? styles.loadingCss : {}}
                >
                  {job.message}
                </Typography>
                <CardActions>
                  {job.status === "done" && (
                    <Button size="small">Download</Button>
                  )}
                  {job.status !== "in_progress" && (
                    <Button size="small" onClick={() => handleDismiss(job.id)}>
                      Dismiss
                    </Button>
                  )}
                </CardActions>
              </CardContent>
            </Card>
          ))}
        </Stack>
      </Stack>
    </Container>
  );
}

export default App;
