// Represents the different sections of connected integrations we display on the /resources page.
export enum ConnectedIntegrationType {
  Compute = 'Compute',
  Data = 'Data',
  // This is currently only used to filter out integrations we don't want to show as data or compute
  // right now (eg. 'Filesystem').
  ArtifactStorage = 'Artifact Storage',
  Other = 'Other',
}
