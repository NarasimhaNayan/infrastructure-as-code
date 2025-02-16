# infrastructure-as-code

Challenges faced / steps to set up the CI/CD pipeline
1. I was not able to use the 'ubuntu-lastest' as the vmImage for the default Azure pipeline pool. This is because of the  : ##[error]No hosted parallelism has been purchased or granted. To request a free parallelism grant, please fill out the following form https://aka.ms/azpipelines-parallelism-request
 Solution : Set up and configured self-hosted agent to run the pipelines.
 Ref: [Stack-overflow](https://stackoverflow.com/questions/68405027/how-to-resolve-no-hosted-parallelism-has-been-purchased-or-granted-in-free-tie)

 [Setting up the self hosted agent](https://learn.microsoft.com/en-us/azure/devops/pipelines/agents/windows-agent?view=azure-devops)
