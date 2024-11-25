# Thela Locator 

## Finding all the active and nearby thelas geographically within 5 kms range.
![image](https://github.com/user-attachments/assets/b97688c1-b92e-4f9a-9f2d-cc2bd6fe1d86)

Follows the following HLD

#
## 
  - Similar to cab finding on OLA/Uber
  - (https://github.com/KinMod-ui/RRloadbalancer) which is a auto scaling/load balancer application which launches thelaLocator server in the backend.
  - The loadbalancer also acts as a reverse-proxy for our initial request to the backend server which passes through loadbalancer to the backend server and the responses travel through it too masking the backend's identity.
  - The future requests run on a persistent websocket connection with the backend to send location and get data of our nearby friends according to location.

## FOR FUTURE REFERENCE   
  - The reverseproxy strips connection header which should have been upgrade to the backend server when getting a websocket connection request from frontend and passing through the proxy to follow RFC guidelines.
  - That is why we weren't able to establish a straight websocket connection between frontend and backend and that would be costly as the communication is always through the middle server to the frontend or backend thereby increasing communication costs.
  - There will be a persistent websocket connection between proxy and backend and one between proxy and client which isn't ideal for us
  - Rather what we did is http request through loadbalancer/proxy to get the backend server through any strategy(here round robin) and then only forming websocket persistent connection with that backend instead of with both.
