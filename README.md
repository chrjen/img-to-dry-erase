# Image to dry erase
Convert textures into .vpcf file to be used with dry erase boards in Half-Life: Alyx

Drop the .exe in the top directory in your Half-Life: Alyx SDK.
Open the directory in the command line and specify the relative path to .vtex texture.
This should automatically create a particle system file (.vpcf) for you.

Use argument `-help` for options.

Example:
`> img-to-dry-erase.exe -vtex="materials/particle/dry_erase/my_texture.vtex"`
