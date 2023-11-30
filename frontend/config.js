const APP_CONFIG = {
  api: {
    host: "localhost",
    port: 8080,
    proto: "http",
  },
};

document.dispatchEvent(new Event("configLoaded"));
