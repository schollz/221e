module.exports = {
  apps: [
    {
      name: "jackd",
      script: "jackd",
      //args: "-d alsa -P hw:0",
      args: "-d dummy",
      //args: "-d alsa -C hw:5 -P hw:5",
      autorestart: true,
      watch: false
    },
    {
      name: "supercollider",
      script: "sclang",
      args: "internal/supercollider/sampler.scd",
      autorestart: true,
      watch: false,
      wait_ready: true,
    }
  ]
};

