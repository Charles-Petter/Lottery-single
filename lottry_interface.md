http://localhost:8081/lottery/v1/get_lucky

```plain
{
   "user_id": 1, 
   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VyX25hbWUiOiJ6aGFuZ3NhbiIsIlN0YW5kYXJkQ2xhaW1zIjp7ImV4cCI6MTcxNDEzMDk0NiwiaXNzIjoibG90dGVyeSIsIm5iZiI6MTcxNDEyMzc0Nn19.UxKlI4M1m4vLrYhiGBx9_A-aGCOW9Hh9RNI__2sfDwc",
   "ip":"192.168.62.1"

}
```

http://localhost:8081/admin/login



```plain
{
 "user_name": "zhangsan",
  "pass_word": "123321"

}
```

http://localhost:8081/admin/register

```plain
{
  "user_name": "admin3",
  "pass_word": "123456",
  "email": "17482255@qq.com",
  "mobile": "15807414130",
  "real_name": "John Doe",
  "age": 30,
  "gender": "male"
}
```

http://localhost:8081/admin/get_prize_list

![img](https://cdn.nlark.com/yuque/0/2024/png/29337569/1714487829256-b7badf1c-b269-4fce-a978-58c6e34abb8b.png)

http://localhost:8081/admin/add_prize

```
{
  "id":6,
  "title":"xiaomi14mi14",
  "img":"https://p0.ssl.qhmsg.com/t016ff98b934914aca6.png",
  "prize_num":10,
  "prize_code":"0-9999",
  "prize_time":0,
  "left_num":0,
  "prize_type":5,
  "prize_plan":"",
  "begin_time":"2024-05-07T13:55:16.6551796+08:00",
  "end_time":"2024-05-14T13:55:16.6551796+08:00",
  "display_order":0,
  "sys_status":0
}
```

![image-20240430224316628](C:\Users\Meet\AppData\Roaming\Typora\typora-user-images\image-20240430224316628.png)

http://localhost:8081/admin/blackip/add

```plain
{
    "ip": "192.168.1.100",
    "blackTime": "2024-05-06T00:00:00Z"
}
```

![img](https://cdn.nlark.com/yuque/0/2024/png/29337569/1714910533144-73817c43-b62d-4fae-bb8c-ec5e467a44d1.png)





http://localhost:8081/admin/blackip/list

![img](https://cdn.nlark.com/yuque/0/2024/png/29337569/1714910599977-3d7f036b-c9b4-41ca-87ae-f0271ab07aa5.png)

http://localhost:8081/admin/blackip/delete/1

![img](https://cdn.nlark.com/yuque/0/2024/png/29337569/1714910755797-6f6883a4-9af8-492c-be0c-b977de241e7c.png)





http://localhost:8081/admin/delete_prize/2

![img](https://cdn.nlark.com/yuque/0/2024/png/29337569/1714913920498-5ff9e8ca-a880-47fc-8237-432c31d414ab.png)