SELECT *
FROM `user`
WHERE `id` IN ({{list .ids}})