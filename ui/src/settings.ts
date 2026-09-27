const available = {
  sources: [
    { value: "ADF Front", label: "Single Sided" },
    { value: "ADF Duplex", label: "Double Sided" },
  ],

  modes: [
    { value: "Color", label: "Color" },
    { value: "Gray", label: "Gray" },
    { value: "Lineart", label: "Monochrome" },
  ],

  resolutions: [150, 200, 300, 400, 600].map((value) => ({
    value,
    label: value.toString(),
  })),
};

const defaults = {
  mode: available.modes[0].value,
  resolution: 300,
  source: available.sources[0].value,
  clean: true,
  pdf: true,
};

export default {
  available,
  defaults,
};
