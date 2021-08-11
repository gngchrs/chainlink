-- Password for all encrypted keys is 'p4SsW0rD1!@#_'
-- Scrypt params are chosen to be completely insecure and very fast to decrypt
-- Don't use any of these keys for anything outside of testing!

INSERT INTO "public"."encrypted_ocr_key_bundles"("id","on_chain_signing_address","off_chain_public_key","encrypted_private_keys","created_at","updated_at","config_public_key")
VALUES
(
    DECODE('7f993fb701b3410b1f6e8d4d93a7462754d24609b9b31a4fe64a0cb475a4d934','hex'),
    DECODE('57120312197c54ce8eeca1aabef8682d3d768bf4','hex'),
    DECODE('e82a081a49d27f11b526c42fe8228837832b25598a4ff146eb173f30a070030e','hex'),
    E'{"kdf": "scrypt", "mac": "01623cb2f2eac87e73b082a159960003ad1d627f0a3e4666bbeb4808baf8e367", "cipher": "aes-128-ctr", "kdfparams": {"n": 2, "p": 1, "r": 8, "salt": "0800e3d8a409beeb74c33ea539994e4c2fa22ef62735fa866bc9e146af1369bf", "dklen": 32}, "ciphertext": "bf931de9605adadebe2e833ea457a4629fa90a4ffff696f869a6f4ee29d4df1df8d2ecf68940d4134f95d70a78a561051d0b578e6c7fda4b0921e445595fe6ca25fa4421b022b1c2b13990476fe7b061c2ebbbbd6d1b1b9f9b9048078f6c0115f63e8d44d118ba82914e24a06ea09d31567f3c21ccac620d2c22f0fd2e1144c3b0ad66072f523096728536e7575fbf2c66c96f5cfce377a45e03bf230758edacbfeba55731fa59b43d8d794a32931a6f44740f7ad56fb6f6a90cd33f6021a6a25b71835bdacf0f5c544412af633d4a752e35ab973659a93154526354cf6404196daf0a9821faf3d231524763fc9dc5cd11291467503bec5d599e5ba1ee040e80a51b624bccbc1f769eaa798627bffb94cdfa12bec7b33dff28fb21a8b0e42f812c0b172b3b51c71d089f9f0d48984e98e030e97e672c6754e0a8fc1d1fc3ea1823894d8f79b5f8fd6c12cfa36b73", "cipherparams": {"iv": "cb05be43c5a44baf33ddba0d62b45017"}}',
    E'2021-01-22 02:59:40.087088+00',
    E'2021-01-22 02:59:40.087088+00',
    DECODE('1bd910cb46ffce56ebd59721e406bd4e7e76f9d0561352703bacc18e165af174','hex')
);

INSERT INTO "public"."encrypted_ocr2_key_bundles"("id","onchain_public_key","onchain_signing_address","offchain_signing_public_key","offchain_encryption_public_key","encrypted_private_keys","created_at","updated_at")
VALUES
(
    DECODE('b0d6de2a01eb2d66b4db2cac7d4aea47c701612a1d0dc892fc59e8e9068b23f6','hex'),
    DECODE('04e951b640045c44cee804b15c5f9d6b9520b50414eb56a33f695f4385032a43fb2240887ce0ea3189bbe185dc28a2e87bfe34ca1bcac115be6259e67905ff2392','hex'),
    DECODE('61f71811e34b5dd891fcb30f8f96855bab9ddafc','hex'),
    DECODE('86c21ddb98a74d779102bd688dd0022839397644286a7d308f096d2320c104f9','hex'),
    DECODE('b5263c71618e298b6798f9a6849f805b54ae3570fbe5d989ef8eb89e20e9d2f9','hex'),
    E'{"kdf": "scrypt", "mac": "98b0908288bc48c0b128dd995a5e780de4164526c13e47ad879751f07fedb98b", "cipher": "aes-128-ctr", "kdfparams": {"n": 2, "p": 1, "r": 8, "salt": "e67a937f07069c118c41ba546b71a505ac418360af4b39d51a0a1bfa5a4401c5", "dklen": 32}, "ciphertext": "d728556f4556937bf24bff67b571ca91bf90d7d1b78ec841b0cc999fac2ecb2bed20c8494c05ed90c4a35677a327689c64d8d913182d33e865ecd20c90d835dc18359c482e4dad9d9fc771cfb97c6dbf30c7937a13a2f99a1f3f54360b5a805d93066ae479597983b67be61a18c70be2903f4fbe1efede28d0705dcf7cf6a9f29170fb63aee87fb34b879d709483a705810ebb553c784af9164d272bf17fb4330522184bf368892a218ec92e0063ea35e332053ad62c2db0678f9bdeb458c510fc0f82a884320ac6856a1962f177bdd1efda838fccb0124e564186beb259f0b0561f9b71e6999956307274f34930f1b8afd2d7a4e8ce621f867e14eb446d9dc5145a", "cipherparams": {"iv": "8c94c6c0aa0fa2821f29e32b9f21a97a"}}',
    E'2021-08-06 17:40:05.219808+00',
    E'2021-08-06 17:40:05.219808+00'
);

INSERT INTO "public"."encrypted_p2p_keys"("peer_id","pub_key","encrypted_priv_key","created_at","updated_at")
VALUES
(
    E'12D3KooWApUJaQB2saFjyEUfq6BmysnsSnhLnY5CF9tURYVKgoXK',
    DECODE('0ee29fb3bbcec807959c3f9e85c43ac85570c8a63d94e444599527884c992ece','hex'),
    E'{"kdf": "scrypt", "mac": "f1fde38d7d0a86c25eb8ce64885db212e10cbb493671e861151adc14367a4b25", "cipher": "aes-128-ctr", "kdfparams": {"n": 2, "p": 1, "r": 8, "salt": "7bfa8fab0a3f3907b1df7479d92183d32610b471c4a96629fa1f05428914b168", "dklen": 32}, "ciphertext": "a99751e60f8c8143b17635c107aa4ac1703ebb4f2eaf52d4e6e037b196c1e319a482d5a8cc8d49198f0de912452e42283132e53b0d931cf10d71bd641691919719144f0a", "cipherparams": {"iv": "88b81ab2a402b1a6c0981cc521ca24fa"}}',
    E'2021-01-22 02:59:40.085609+00',
    E'2021-01-22 02:59:40.085609+00'
);

INSERT INTO users (email, hashed_password, token_secret, created_at, updated_at) VALUES (
    'apiuser@chainlink.test',
    '$2a$10$Ee8YjCtcBgflgR7NWmii.u5kwOuWNF1bniacRf/sqobB5YaQv.Lm.', -- hash of literal string 'p4SsW0rD1!@#_'
    '1eCP/w0llVkchejFaoBpfIGaLRxZK54lTXBCT22YLW+pdzE4Fafy/XO5LoJ2uwHi',
    '2019-01-01',
    '2019-01-01'
);
