# Thela Locator 

## Finding all the nearby thelas geographically within 5 kms range.
![image](https://github.com/user-attachments/assets/b97688c1-b92e-4f9a-9f2d-cc2bd6fe1d86)

Follows the following HLD

#
## 
  - (https://github.com/KinMod-ui/RRloadbalancer) which is a auto scaling/load balancer application which launches thelaLocator server in the backend.
  - The loadbalancer also acts as a reverse-proxy for our initial request to the backend server which passes through loadbalancer to the backend server and the responses travel through it too masking the backend's identity.
  - The future requests run on a persistent websocket connection with the backend to send location and get data of our nearby friends according to location.
