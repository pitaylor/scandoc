import {useState} from "react";

export const Settings = {
  sources: [
    {value: "ADF Front", label: "Single Sided"},
    {value: "ADF Duplex", label: "Double Sided"},
  ],

  modes: [
    {value: "Color", label: "Color"},
    {value: "Gray", label: "Gray"},
    {value: "Lineart", label: "Monochrome"},
  ],

  resolutions: [150, 200, 300, 400, 600].map((value) => ({value, label: value.toString()}))
};

export const useParams = () => {
  const [params, setParams] = useState({
    mode: Settings.modes[0].value,
    resolution: 300,
    source: Settings.sources[0].value,
    clean: true,
    pdf: true
  });

  const {mode, resolution, source, clean, pdf} = params;

  const searchParams = new URLSearchParams({
    mode,
    source,
    resolution: resolution.toString(),
    clean: clean.toString(),
    pdf: pdf.toString()
  });

  return {params, searchParams, setParams};
};