# Public Inputs Format (v1)

The stub and circuit share the same public inputs ordering.

Encoding:
`noema_public_inputs_v1|pt=<int>|ms=<int>|op=<0|1>|c=<hex>`

Fields:
- `pt`: policy threshold (0..2)
- `ms`: max severity (0..2)
- `op`: overall pass (1) or fail (0)
- `c`: commitment hex string with `0x` prefix

`public_inputs_b64` is the base64 encoding of the UTF-8 bytes above.
`proof_b64` is a stub proof derived from the public inputs.
