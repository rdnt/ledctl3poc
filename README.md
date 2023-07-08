# ledctl3poc

This is a proof-of-concept rebuild of the architecture of [ledctl](https://github.com/rdnt/ledctl), and it will hopefully be moved to the main repo once it is a viable implementation.

Its goals include but are not limited to:

- Multiple sources: multiple source devices with multitudes of capabilities and input drivers (e.g. video capture, audio capture, effects engines)
- Multiple sinks: multiple sink devices with an array of outputs for controlling separate LED strips from the same device.
- Centralized configuration: configuration registry will be a single entry point for discovering devices, configuring device parameters like led count & led calibration, and setting up and triggering profiles that individually and in-parallel control the device mesh.
- Lightweight: pluggable networking (main implementation will be JSON over websockets), allowing easier testing of the protocol and using different implementations.
- Ability to ember effects engines (sources) and renderers (sinks) into the same binary, allowing for server-rendered effects for better performance and ease of use.

I am currently developing this architecture in my free time and in the open. If you are interesting in contributing/collaborating, please reach out to me via Discord (rdnt).
