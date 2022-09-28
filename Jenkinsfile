@Library('jenkins-library@opensource-release') _
dockerImagePipeline(
  script: this,
  service: 'dns',
  dockerfile: 'Dockerfile',
  buildContext: '.',
  buildArguments: [PLATFORM:"amd64"]
  
)
