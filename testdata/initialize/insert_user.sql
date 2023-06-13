/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

INSERT INTO `user`
    (`tenant_id`,`name`, `password`, `birthday`, `address`, `role`)
values (:tenant_id,:name, :password, :birthday, :address, :role)