# istio-content-based-routing
Demonstrating how to use Istio to route to applications in different namespaces using Cookies or JWT Claims.

## Overview
Sometimes it can be handy to deploy multiple versions of an app concurrently, and have the version accessed be determined based on the request content.

Istio allows content based routing, and is able to perform that based upon http headers (among other factors), but not JWT claims, nor cookies specifically. Thankfully JWTs and Cookies are both http headers, so there's a way to make this work. 

## Cookie
Istio can route based on Cookies relatively easily, although we have to use a regular expression to extract the cookie name & value from the overall cookie header. Here's the one used in the example, looking for a cookie called `Istio-NS-Hint` with a value of `test1`

```yaml
  - match:
    - headers:
        cookie:
          regex: (^|.*; )Istio-NS-Hint=test1($|; .*)
```          

To set the cookie, we can map in a small application that will set the routing cookie based on the url invoked. A simple application to do this, written in Go, is in the cookiesetter folder of this repo, along with a Dockerfile to build it, alternatively, use the prebuilt version in dockerhub.

### JWT
Istio does not offer support for routing based on claims within a JWT, but we can achieve this functionally by using an Istio Envoy Filter to read the JWT, and republish the claims from it as HTTP headers. Then Istio can be configured to route based on the new headers. 

The full code for the filter is in the `istio-envoy-filter-jwt-lua.yaml` file.

With the filter in place, we can route based on identity, or other claims within the jwt (such as group memberships, or audience/scope).

## Try it out..

Bring up a cluster with istio.. and deploy (via `kubectl apply -f <filename>`)

- `headers-test1.yaml` 
  - a simple container that echoes http headers back to the requestor, deploys to namespace `test1`
- `headers-test2.yaml` 
  - a different simple container that also echoes headers back (with slightly different output). Deploys to namespace `test2`
- `cookiesetter/cookiesetter.yaml`
  - a very small rest endpoint that sets a `Istio-NS-Hint` cookie with the value from the url path, eg visiting `/cookie/test` will set the cookie with the value `test1`. See code in the `cookiesetter` folder for more info, Dockerfile provided.
- `istio-envoy-filter-jwt-lua.yaml` a simple lua filter that base64 decodes the jwt, decodes the json, then sets headers based on the claims.
- `hello-istio-gateway.yaml` 
  - an istio gateway & virtual service that maps `/cookie/<cookie-value>` to a cookiesetting application, and all other requests are tested for; 
    - a jwt (present as the value for the http header `jwtheadername`) with a `name` claim of `John Doe`. 
    - a cookie with the name `Istio-NS-Hint` and the value `test1` 
  - If either condition is met, the request is routed to the `test1` headers application, else the request flows to the `test2` namespace
  
Access the application by looking up the nodeport for the istio ingress, and visiting it via http, or a rest client. 
jwt.io can be used to generate sample jwts to place in a header called `jwtheadername`.

## Awesome! is this production ready ?

Firstly note that all these approaches do is allow the routing to be selected, this isn't supposed to be any form of security. The destination application (or other layer) is still required to ensure the request is allowed or not. This is all about selecting which app to invoke, not enforcing that selection. (Eg. note that the JWT parsing doesn't even verify the signature, as that's assumed to be handled by the application itself). It's not difficult for users to craft requests with cookies in, or with jwts. Keep this in mind and design accordingly. 

That said, the Cookie approach, sure, it would work fine in production, it can be real handy as a quick way to switch between versions of apps during demos =)

The JWT approach, comes with some larger caveats. The filter is written in LUA, and brings along its own rudimentary base64 decoder, and json decoder. The base64 decoder is far from optimal, the json decoder is fair, but again, if performance is a concern (remember this filter runs at the gateway), then you may wish to swap the implementations out for better ones. The ones used here are selected because they were compatible with Apache licence for this project. Also, note that while this works fine for simple claims of String or Number, if a claim is an array, or a json Object, then the challenges become a little harder. The current filter is written to try to enable these usages (you can see the headers generated in the echo applications), but ymmv.

## Is there a better way ?

Probably, Mixer Adapters look like they may be a lot better suited to handling this, but they are only in prereleases of Istio at this time.

