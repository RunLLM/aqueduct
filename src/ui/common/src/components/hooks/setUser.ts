import cookie from 'react-cookies';

export default function setUser(apiKey: string): void {
  cookie.save('aqueduct-api-key', apiKey, { path: '/' });
}
