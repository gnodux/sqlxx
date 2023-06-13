SELECT * FROM `user`
WHERE `role` IN ({{list .roles}})