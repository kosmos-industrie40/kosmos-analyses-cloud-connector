# Test

In this file, you can find a description, how to test different endpoints.

## Content

1. [Auth](#auth)

## Auth
In this chapter a simple test case against the auth endpoint can be found. In three steps we will try to logged in, test if we are already logged in and log out.
We are using `curl` to send the requests.

### Log in
```bash
curl -i -X POST localhost:8080 --data '{"user":"test", "password":"abc"}'
```

### Test Authenticated
```bash
curl -i --header 'token:(RETURN VALUE FROM LOG IN REQUEST)' localhost:8080 
```

### Logout
```bash
curl -i -X DELETE --header 'token:(RETURN VALUE FROM LOG IN REQUEST)' localhost:8080 
```


