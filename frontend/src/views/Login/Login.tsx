import React, { useState } from 'react';
import axios from 'axios';
import Button from '@material-ui/core/Button';
import CircularProgress from '@material-ui/core/CircularProgress';
import Container from '@material-ui/core/Container';
import Paper from '@material-ui/core/Paper';
import TextField from '@material-ui/core/TextField';
import Typography from '@material-ui/core/Typography';
import { useHistory } from 'react-router-dom';

const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const history = useHistory();

  const submit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    setErrorMessage('');
    axios
      .post<{ redirect: string }>('/api/ldap/login', { username, password })
      .then(({ data }) => {
        setPassword('');
        history.replace(data.redirect || '/');
      })
      .catch((error) => {
        setPassword('');
        if (error.response?.status === 401) {
          setErrorMessage('Invalid username or password.');
        } else if (error.response?.status === 403) {
          setErrorMessage('You are not authorized to access this application.');
        } else {
          setErrorMessage('Sign-in is unavailable. Please try again.');
        }
      })
      .finally(() => setSubmitting(false));
  };

  return (
    <Container maxWidth="xs" style={{ paddingTop: 96 }}>
      <Paper elevation={3} style={{ padding: 32 }}>
        <Typography variant="h5" component="h1" gutterBottom>
          Sign in
        </Typography>
        <form onSubmit={submit}>
          <TextField
            autoFocus
            autoComplete="username"
            disabled={submitting}
            fullWidth
            label="Username"
            margin="normal"
            onChange={(event) => setUsername(event.target.value)}
            required
            value={username}
            variant="outlined"
          />
          <TextField
            autoComplete="current-password"
            disabled={submitting}
            fullWidth
            label="Password"
            margin="normal"
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
            variant="outlined"
          />
          {errorMessage && (
            <Typography color="error" role="alert" variant="body2">
              {errorMessage}
            </Typography>
          )}
          <Button
            color="primary"
            disabled={submitting}
            fullWidth
            type="submit"
            variant="contained"
          >
            {submitting ? <CircularProgress size={20} /> : 'Sign in'}
          </Button>
        </form>
      </Paper>
    </Container>
  );
};

export default Login;
