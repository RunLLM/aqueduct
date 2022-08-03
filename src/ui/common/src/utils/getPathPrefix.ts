// `getPathPrefix` is a helper function that inspects the current pathname and
// determines any proxy prefixes that might be present. If the proxy prefix is
// present, then `getPathPrefix` returns a string that represents the prefix
// that should be appended to all routes.
//
// If no prefix is present, `getPathPrefix` returns an empty string.
export const getPathPrefix = () => {
  const location = window && window.location ? window.location.pathname : '';

  const result = location.match(SageMakerProxyRegex);
  if (!result) {
    return '';
  } else {
    return result[0]; // The regex that was matched.
  }
};

const SageMakerProxyRegex = /\/proxy\/\d{4}/;
