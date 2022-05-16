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
  Typography,
} from "@mui/material";
import { Settings, useParams } from "./Settings";

const WS_HOST =
  process.env.NODE_ENV === "development"
    ? "localhost:8090"
    : document.location.host;

interface Job {
  id: string;
  name: string;
  status?: string;
  message?: string;
}

const fileName = (fn: string) => {
  const i = fn.lastIndexOf("/");
  if (i !== -1) {
    return fn.slice(i + 1);
  } else {
    return fn;
  }
};

function App() {
  const { modes, resolutions, sources } = Settings;
  const { params, searchParams, setParams } = useParams();
  const [error, setError] = useState("");
  const [scanning, setScanning] = useState(false);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const input = event.target as HTMLInputElement;
    const value = input.type === "checkbox" ? input.checked : input.value;
    setParams({ ...params, [input.name]: value });
  };

  const ws = useRef<null | WebSocket>(null);
  const outgoing = useRef<string[]>([]);
  const jobsRef = useRef<Record<string, Job>>({});
  const [sentinel, setSentinel] = useState(0);
  const [jobs, setJobs] = useState<Job[]>([]);

  const sendRequests = useRef(() => {
    if (ws.current == null) {
      return;
    }
    if (ws.current.readyState >= 2) {
      ws.current = null;
      setSentinel(sentinel + 1);
      return;
    }
    let request = outgoing.current.shift();
    while (request) {
      ws.current.send(request);
      request = outgoing.current.shift();
    }
  });

  const sendRequest = (request: string) => {
    outgoing.current.push(request);
    sendRequests.current();
  };

  const blitJobs = () => {
    setJobs(Object.values(jobsRef.current).reverse());
  };

  useEffect(() => {
    ws.current = new WebSocket(`ws://${WS_HOST}/ws`);
    ws.current.onopen = () => {
      console.log("websocket opened", sentinel);
      sendRequests.current();
    };
    ws.current.onclose = () => console.log("websocket closed");
    ws.current.onmessage = (event) => {
      const response = JSON.parse(event.data);
      const { job, status, message } = response;
      if (job) {
        const { id, name } = job;
        jobsRef.current[id] = { id, name, status, message };
        if (message.includes("scanning done")) {
          setScanning(false);
        }
      } else if (status === "failed") {
        setError(message);
        setScanning(false);
      }
    };

    const wsCurrent = ws.current;
    const intervalId = setInterval(blitJobs, 250);

    return () => {
      if (wsCurrent != null) {
        wsCurrent.close();
      }
      clearInterval(intervalId);
    };
  }, [sentinel]);

  const handleScan = () => {
    setError("");
    setScanning(true);
    sendRequest(searchParams.toString());
  };

  const handleDismiss = (jobId: string) => {
    delete jobsRef.current[jobId];
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
                value={params.source}
                name="source"
                onChange={handleChange}
                row
              >
                {sources.map(({ value, label }) => (
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
                value={params.mode}
                name="mode"
                onChange={handleChange}
                row
              >
                {modes.map(({ value, label }) => (
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
                  min={resolutions[0].value}
                  max={resolutions[resolutions.length - 1].value}
                  value={params.resolution}
                  onChange={(event: Event, value: number | number[]) =>
                    setParams({
                      ...params,
                      resolution: value as number,
                    })
                  }
                  step={null}
                  marks={resolutions}
                />
              </Box>
            </FormControl>

            <FormControlLabel
              control={
                <Switch
                  name={"clean"}
                  checked={params.clean}
                  onChange={handleChange}
                />
              }
              label="Auto Clean"
            />

            <FormControlLabel
              control={
                <Switch
                  name={"pdf"}
                  checked={params.pdf}
                  onChange={handleChange}
                />
              }
              label="PDF"
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
                <Typography variant="body2">{fileName(job.name)}</Typography>
                <Typography variant="body2" color="text.secondary">
                  {job.message}
                </Typography>
                <CardActions>
                  {job.status === "done" && (
                    <React.Fragment>
                      <Button size="small">Download</Button>
                      <Button size="small">Delete</Button>
                    </React.Fragment>
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
