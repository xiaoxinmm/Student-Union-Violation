# 程序说明

1.连接不上数据库 请检查 php . ini中的mysql扩展是否开启。检查php下的ext文件夹中是否有mysql扩展dll

2.检查 mysql.php是否为你当前数据库 建议 数据名改为db1 表则自动生成

3.代码需优化ing~~

4.保障安全的话建议在php.ini设置一个时区！

data.timezone = PRC

5.文件deletuser.php是个页面可以封装成函数 后期可能！

6.生成本地文件 这里可能用python 生成word文档 excel可以直接导出

7.图片查看系统是有ftp驱动的 你可以去指定个文件夹 或者到 aydutsys里面修改查看图片那部分，主页面的图查系统是连接ftp按钮，根据需求可以删除。

8.这里的公示系统不建议改因为已经配置好pythonscript脚本了，所以尽量不要改，脚本是生成excel表格操作留作本地保存。

9.db1数据库预先创建好，表内容可以用sql导入进去。用户名和学生科用户明已经安排好了

10.账号密码自己看数据表md5 自己解密一下
