export default {
  loadingCss: {
    "&:after": {
      overflow: "hidden",
      display: "inline-block",
      verticalAlign: "bottom",
      animation: "ellipsis steps(4,end) 900ms infinite",
      content: '"\u2026"',
      width: 0,
    },
    "@keyframes ellipsis": {
      to: {
        width: "1.25em",
      },
    },
  },
};
