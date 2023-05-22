// Represents the different sections of connected resources we display on the /resources page.
export enum ConnectedResourceType {
  Compute = 'Compute',
  Data = 'Data',
  // This is currently only used to filter out resources we don't want to show as data or compute
  // right now (eg. 'Filesystem').
  ArtifactStorage = 'Artifact Storage',
  Other = 'Other',
}
