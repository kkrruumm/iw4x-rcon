# iw4x-rcon
an rcon client for iw4x

# Usage
* Safe-RCON:

    - `./iw4x-rcon -i <ip_of_server> -k <path_to_private_key>`

* Insecure-RCON:

    - `./iw4x-rcon -i <ip_of_server> -pass <RCON_passphrase>`

If you need to specify the port of the server you are connecting to, you may append the `-p` option. Example:

   - `./iw4x-rcon -i <ip_of_server> -k <path_to_private_key> -p 28965`

# Building

`go` is required to build iw4x-rcon.

1. Clone the source, move to its directory.
2. `go build -ldflags="-s -w"` in the source directory. `ldflags` here strips the binary.
3. After building, you should have a resulting `iw4x-rcon` binary, done!

# Running

Since this produces a single binary, you can just start it with `./` on *nix systems or it can be added to $PATH somewhere. Typically, on Linux a good location is `/usr/local/bin`.

# Misc

I highly recommend using safe RCON, as insecure RCON broadcasts your RCON passphrase in plaintext across the internet along with your command- meaning that anyone who manages to intercept it will have gained control of your server.

The expected keypair is RSA 4096, and with openssl on *nix systems, can be generated like so:

1. `openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:4096`
2. `openssl rsa -pubout -in private_key.pem -out public_key.pem`
3. `openssl rsa -pubin -inform PEM -in public_key.pem -outform DER -out rsa-public.key`

Once you have your `rsa-public.key`, you may copy that to your IW4x server and place it *immediately* next to the `iw4x.exe` binary- the server should automatically pick it up.

**You should not** copy your *private* key to the server, and instead keep that on the computer you are going to be connecting to the server from. This is the value that the `-k` argument expects. Be sure to not share your *private* key, as is how the server authenticates you.