import React, {useState} from 'react';
import {
  Alert,
  Box,
  Button,
  Container,
  FormControl,
  FormControlLabel,
  FormGroup,
  FormLabel,
  Radio,
  RadioGroup,
  Slider, Stack,
  Switch,
} from "@mui/material";
import {Settings, useParams} from "./Settings";

function App() {
  const {modes, resolutions, sources} = Settings;
  const {params, searchParams, setParams} = useParams();
  const [error, setError] = useState("");
  const [submitting, setSubmit] = useState(false);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const input = event.target as HTMLInputElement;
    const value = input.type === "checkbox" ? input.checked : input.value;
    setParams({...params, [input.name]: value});
  };

  const handleSubmit = () => {
    setError("");
    setSubmit(true);

    fetch("?" + searchParams, {method: "POST"})
      .then(response => {
        if (response.ok) {
          setTimeout(() => setSubmit(false), 1500);
        } else {
          setError(response.statusText);
          setSubmit(false);
        }
      })
      .catch(() => {
        setError("An unexpected error occurred");
        setSubmit(false);
      });
  };

  return (
    <Container maxWidth="xs">
      <FormGroup sx={{marginTop: '16px'}}>
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
              {sources.map(({value, label}) =>
                <FormControlLabel key={value} value={value} label={label} control={<Radio/>}/>)}
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
              {modes.map(({value, label}) =>
                <FormControlLabel key={value} value={value} label={label} control={<Radio/>}/>)}
            </RadioGroup>
          </FormControl>

          <FormControl>
            <FormLabel id="resolution-label">Resolution</FormLabel>
            <Box sx={{margin: "0 10px"}}>
              <Slider
                aria-labelledby="resolution-label"
                min={resolutions[0].value}
                max={resolutions[resolutions.length - 1].value}
                value={params.resolution}
                onChange={(event: Event, value: number | number[]) => setParams({
                  ...params,
                  resolution: value as number
                })}
                step={null}
                marks={resolutions}
              />
            </Box>
          </FormControl>

          <FormControlLabel
            control={<Switch name={"clean"} checked={params.clean} onChange={handleChange}/>}
            label="Auto Clean"
          />

          <FormControlLabel
            control={<Switch name={"pdf"} checked={params.pdf} onChange={handleChange}/>}
            label="PDF"
          />

          <Button variant="contained" disabled={submitting} onClick={handleSubmit}>
            {submitting ? 'Submitting...' : 'Scan'}
          </Button>

          {error && <Alert onClose={() => setError("")} severity="error">{error}</Alert>}
        </Stack>
      </FormGroup>
    </Container>
  );
}

export default App;
