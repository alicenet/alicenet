package transport

// TODO: expand docs

/*
This package uses the lightning network Brontide protocol to create an
authenticated and encrypted transport based on the Noise Protocol Framework.
In order to isolate the external dependencies from the remainder of the project,
all externally referenced packages have been converted into interfaces.

Note: Brontide enforces a 5 second timeout
*/
