# Contribution Guidelines

First of all, thanks for taking your time to contribute to the AliceNet development, you are awesome!

However, before your submit your cool code, consider the following guidelines:

Search on GitHub for an open or closed PR that relates to your submission. You don't want to duplicate existing efforts.

Be sure that an issue describes the problem you're fixing, or documents the design for the feature you'd like to add. Discussing the design upfront helps to ensure that we're ready to accept your work.

Also, read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) to help us maintain the AliceNet development community a healthy environment.

Now that you are ready, the following sections will help you to get your work ready for submission.

## Forking and cloning the repository

The first step to help on the AliceNet development is to fork the AliceNet repository (ideally using SSH). See the official github documentation for details on how to set up your [ssh credentials](https://docs.github.com/en/authentication/connecting-to-github-with-ssh) and how to [fork a repository](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository).

Once you have created your fork, open a new terminal and run the following commands:

```shell
git clone --recursive git@github.com:[username]/alicenet.git
cd alicenet
git remote add upstream git@github.com:alicenet/alicenet.git
git fetch upstream
```

## Building the repository from source code

Once you have you repository cloned, refer to the [Building Documentation](./docs/BUILD.md) for more information on to build the binary from the source code. Now, you can start to code, and when you have a binary compiling, proceed to the next step.

## Testing

The last step, before creating a Pull Request to submit your work, is to make sure that all tests and linters are passing. Check [How to run the unit tests and linters](./docs/TESTING.md) documentation for more information.

## Submitting your work as a Pull Request

Now, its finally time to submit your work. See the [github pull request documentation](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request-from-a-fork) to see how to create a Pull Request from your fork against the AliceNet repository.

And again, thank you for your help!
