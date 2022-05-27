import cookie from 'react-cookies';

export default function setUser(apiKey: string) {
    cookie.save('aqueduct-api-key', apiKey, { path: '/' });
}
