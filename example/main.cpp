#include <iostream>
#include <stdio.h>
#include "librmq.h"
#include <string.h>
#include <thread>

using namespace std;


void check_conn(GoUintptr conn) {
    while (true) {
        this_thread::sleep_for(1s);
        cout << ((bool)Connected(conn)?"Connected":"Not connected") << endl;
    }
}

int main() {
    // g++ -pthread  main.cpp -o test ./librmq.so && ./test

    auto path = "/home/inf/log/cmn.log";
    GoString file {p:path,strlen(path)};
    auto log = InitLog(file);

    if (log) {
        std::cout << "init" << std::endl;
        PrintLog(GoString{"test1",5});    
    } 
    else 
    {
        std::cout << "log failed" << std::endl;
    }

    auto url = "amqp://guest:guest@localhost:5672/";
    auto conn = Connect(GoString{url,strlen(url)}, 30);
    if (!conn) {
        std::cout <<"failed connection" << std::endl;
        return 0;
    }
    
    thread chConn(check_conn,conn);
  
    
    auto channel = NewChannel(conn);

    std::cout << channel << " "<< conn << std::endl;

    if (!channel) {
        std::cout <<"failed channel" << std::endl;
        return 0;
    }


    GoString exchange{"logs",4};
    GoString kind{"fanout",6};
    GoString queue{"my.pl",5};
    if (!(bool)ExchangeDeclare(channel,exchange,kind,true,false,false,false,NULL)) {
        std::cout <<"failed ExchangeDeclare" << std::endl;
    }

    auto args = MapArgs();
    GoString deadparam{"x-dead-letter-exchange",strlen("x-dead-letter-exchange")};
    GoString deadExchange{"my.dead.topic",strlen("my.dead.topic")};
    MapArgsAdd(args,deadparam,deadExchange);

    if (!(bool)QueueDeclare(channel,queue,true,false,false,false,args)) {
        std::cout <<"failed QueueDeclare" << std::endl;
    }

    FreeObject(args);

    if (!(bool)QueueBind(channel,queue,GoString{},exchange,false,NULL)) {
        std::cout <<"failed QueueBind" << std::endl;
    }


    int i=1000;
    while(--i) {
        string str = R"({"provider":"sailplay","user_phone":"79100000001","order_num":"test_order_2","user_id":1,"cart":{"1":{"sku":"test1", "price":1000.1, "quantity":2}}})";
        
        
        GoSlice data{(void*)str.c_str(),str.size(),str.size()};
        GoString messageID{"123",3};
        if (!(bool)Publish(channel,exchange, GoString{}, false,false,messageID,data)){
            std::cout <<"failed Publish" << std::endl;
            i = 1;
        };
        std::cout << "send:" << i << endl;
        this_thread::sleep_for(100000ms);
    }
    
        this_thread::sleep_for(5000ms);
        
        
        chConn.detach();

    CloseChannel(channel);
    Disconnect(conn);

    FreeObject(channel);
    FreeObject(conn);
    CloseLog();

    std::cout << "Done." << std::endl;
}