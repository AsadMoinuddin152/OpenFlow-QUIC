# OpenFlow-QUIC with EPM
OpenFlow-QUIC with Extend Performance Modular

## Overview of OpenFlow-QUIC with EPM in Go
**OpenFlow-QUIC with EPM** is an extend performance modular implementation of the OpenFlow protocol in Go, serving [Open vSwitch](https://www.openvswitch.org/) and [RYU controller](https://github.com/faucetsdn/ryu) by [quic-go](https://github.com/lucas-clemente/quic-go) v0.22.0 . 

It uses Unix Domain Socket to insert a slight component into OVS and RYU. 

This project provides improved performance for networks using OpenFlow through the use of multiple streams in QUIC and some behaviors based on OpenFlow messages (faster network convergence and reduced overhead for control plane flow table updates).

## INSTALLATION
*Be Careful, this implementation based on quic-go v0.22.0. Some interfaces have been changed in the new quic-go version. You should change the codes in folder "quic-go" and implement them again by yourself.*

- There are examples in ```/ovs-changed```, ```/ryu-changed```, and ```/quic-changed```.

- Compile the dynamic library by CGO
    1. some interfaces are added for supporting EPM to extract the queuing information in the folder of ```/quic-changed/```. 
    2. download the ```/client``` and ```/server```.
    3. run ```go build --buildmode=c-shared -o "yourlibrary.so" *.go``` in these folders, individually.
    4. set their paths into OVS and RYU.

- As an OVS 
    1. download the OVS and run ```./configure```
    2. make sure that you need to change the MakeFile. Add the ```-ldl``` in the line with ```LIBS```. 
    3. change the functions and add the dynamic library loader and library path to support OpenFlow-QUIC with EPM in ```/ovs-master/stream-fd.c``` and ```/ovs-master/stream-tcp.c```. 
    - **If possible, you should create a series of codes to divide QUIC and TCP.**
    4. please add the UDP flow entry into ```/ofproto/in-band.c```.
    5. after preparing, it's time to install OVS by ```./make install```.
    

- As an RYU controller
    1. download the RYU
    2. change the functions and add the dynamic library loader and library path to support OpenFlow-QUIC with EPM in ```/ryu-master/ryu/controller/controller.py``` and ```/ryu-master/ryu/lib/hub.py```. 
    3. install the RYU by ```pip3 install .```.

## Usage
- for OVS
  - run ```./runovs "your switch name" "your switch ip address and mask" "controller address" ```. It will automatically create OVSDB and run OVSDB.
  - you should add the switch port by yourself. There is an example in ```./runovs```.

- for RYU
  - run ```./ryu-manager "your controller application" --ofp-quic-unix-listen" "your unix socket path```. ```your unix socket path``` default values is ```/tmp/ryu_quic.sock```.

## Development
- ```Algorithm.go``` contains the algorithm of EPM
- ```export_to_server.go``` is for CGO compiling a dynamic library, providing the initiation function.
- ```OFHandlerOperatorReceiver.go``` and ```OFHandlerOperatorSender.go``` are responsible for each OpenFlow message executing some strategies.
- **some codes are ready for supporting [multicast QUIC (Multicast Extension for QUIC draft-jholland-quic-multicast-02)](https://datatracker.ietf.org/doc/draft-jholland-quic-multicast/)**

- for client
  - ```QuicClientManager``` controls ```QuicClient```. The ```QuicClient``` creates the connection to QuicServer and returns the object ```QuicServerConnection```.
  - Then, ```QuicServerConnection``` hands over the connection to "ConnectionController", which executes the scheduling by EPM's components.

- for server
  - ```QuicServerManager``` controls ```QuicServer```. The ```QuicServer``` creates the connection to QuicClient and returns the object ```QuicClientConnection```.
  - Then, ```QuicClientConnection``` hands over the connection to "ConnectionController", which executes the scheduling by EPM's components.



## License
Welcome someone who finds this code useful in your research.
The paper is still under submission.

Copyright (c) 2022 AntLab
Licensed under the GPL-3.0 license.