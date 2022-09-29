<div align="center">
	<p>
		<img alt="Thoughtworks Logo" src="https://raw.githubusercontent.com/ThoughtWorks-DPS/static/master/thoughtworks_flamingo_wave.png?sanitize=true" width=200 />
    <br />
		<img alt="DPS Title" src="https://raw.githubusercontent.com/ThoughtWorks-DPS/static/master/EMPCPlatformStarterKitsImage.png?sanitize=true" width=350/>
	</p>
  <h3>sidecar-mutatingwebhook-init-container</h3>
    <a href="https://app.circleci.com/pipelines/github/ThoughtWorks-DPS/sidecar-mutatingwebhook-init-container"><img src="https://circleci.com/gh/ThoughtWorks-DPS/sidecar-mutatingwebhook-init-container.svg?style=shield"></a> <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
</div>
<br />

Init container for managing the deployment of a MutatingWebhookConfiguration to trigger an admission-contoller based on deployments to a namespace based on matching namespace annotation.  

## Usage

This init container is optimized to work in conjuction with a certificate init container, however it can be used independently so long as the required certificate is available on a (configurable) mount path location.  

