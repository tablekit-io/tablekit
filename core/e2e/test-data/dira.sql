-- MySQL dump 10.13  Distrib 8.4.10, for Linux (aarch64)
--
-- Host: localhost    Database: dbctx_test_dira
-- ------------------------------------------------------
-- Server version	8.4.10

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `dbctx_test_dira`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `dbctx_test_dira` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;

USE `dbctx_test_dira`;

--
-- Table structure for table `activity_log`
--

DROP TABLE IF EXISTS `activity_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `activity_log` (
  `id` char(36) NOT NULL,
  `workspace_id` char(36) NOT NULL,
  `actor_id` char(36) NOT NULL COMMENT 'References users.id',
  `entity_type` enum('task','project','sprint','comment') NOT NULL COMMENT 'Polymorphic discriminator; pair with entity_id (NOT a joinable FK)',
  `entity_id` varchar(64) NOT NULL COMMENT 'Polymorphic id; tasks store bigint as string, others store uuid. NOT a joinable FK',
  `action` enum('created','updated','deleted','assigned','status_changed','commented','mentioned','attached') NOT NULL,
  `snapshot_title` varchar(255) DEFAULT NULL COMMENT 'Denormalized copy of tasks.title / projects.name / sprints.name at the time of the event',
  `metadata` json NOT NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `idx_activity_workspace` (`workspace_id`),
  KEY `idx_activity_actor` (`actor_id`),
  KEY `idx_activity_entity` (`entity_type`,`entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `activity_log`
--

LOCK TABLES `activity_log` WRITE;
/*!40000 ALTER TABLE `activity_log` DISABLE KEYS */;
INSERT INTO `activity_log` VALUES ('019f09e1-7e46-7393-a0a1-68bec9d8bddb','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','1','created','Dashboard MVP shell','{}','2026-05-02 01:00:00.000000'),('019f09e1-7e46-7393-a0a1-68bfedf432a1','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','1','status_changed','Dashboard MVP shell','{}','2026-05-03 02:00:00.000000'),('019f09e1-7e46-7393-a0a1-68c0d4409e51','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','1','assigned','Dashboard MVP shell','{}','2026-05-04 03:00:00.000000'),('019f09e1-7e47-7858-8148-257045e9a897','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','1','updated','Dashboard MVP shell','{}','2026-05-05 04:00:00.000000'),('019f09e1-7e47-7858-8148-2571611a596c','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','1','commented','Dashboard MVP shell','{}','2026-05-06 05:00:00.000000'),('019f09e1-7e47-7858-8148-2572ae8e6492','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','2','created','Token-based auth','{}','2026-05-03 02:00:00.000000'),('019f09e1-7e48-7f64-a824-2fc3aa2a2459','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','2','status_changed','Token-based auth','{}','2026-05-04 03:00:00.000000'),('019f09e1-7e48-7f64-a824-2fc4d5b76b40','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','2','assigned','Token-based auth','{}','2026-05-05 04:00:00.000000'),('019f09e1-7e48-7f64-a824-2fc55512b3a0','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','2','updated','Token-based auth','{}','2026-05-06 05:00:00.000000'),('019f09e1-7e49-727b-b833-01e7f7e6e71c','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','2','commented','Token-based auth','{}','2026-05-07 06:00:00.000000'),('019f09e1-7e49-727b-b833-01e87a08bdd2','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','3','created','Hook up Clerk session','{}','2026-05-04 03:00:00.000000'),('019f09e1-7e49-727b-b833-01e9bf1a43ea','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','3','status_changed','Hook up Clerk session','{}','2026-05-05 04:00:00.000000'),('019f09e1-7e4a-78c6-be32-14ec3481c466','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','3','assigned','Hook up Clerk session','{}','2026-05-06 05:00:00.000000'),('019f09e1-7e4a-78c6-be32-14ed66cf02f4','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','3','updated','Hook up Clerk session','{}','2026-05-07 06:00:00.000000'),('019f09e1-7e4a-78c6-be32-14ee5dbdde85','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','3','commented','Hook up Clerk session','{}','2026-05-08 07:00:00.000000'),('019f09e1-7e4b-7464-8cb9-18d92c6f8740','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','4','created','Refresh-token rotation','{}','2026-05-05 04:00:00.000000'),('019f09e1-7e4b-7464-8cb9-18daef4dd5a4','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','4','status_changed','Refresh-token rotation','{}','2026-05-06 05:00:00.000000'),('019f09e1-7e4b-7464-8cb9-18db2b783c6d','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','4','assigned','Refresh-token rotation','{}','2026-05-07 06:00:00.000000'),('019f09e1-7e4b-7464-8cb9-18dce723e6a3','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','4','updated','Refresh-token rotation','{}','2026-05-08 07:00:00.000000'),('019f09e1-7e4c-743b-9572-9ee974eb3c4f','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','4','commented','Refresh-token rotation','{}','2026-05-09 08:00:00.000000'),('019f09e1-7e4c-743b-9572-9eea0727e0d5','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','5','created','Stats widgets','{}','2026-05-06 05:00:00.000000'),('019f09e1-7e4d-7546-96e7-f5847f13b98e','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','5','status_changed','Stats widgets','{}','2026-05-07 06:00:00.000000'),('019f09e1-7e4d-7546-96e7-f585e0d7e55e','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','5','assigned','Stats widgets','{}','2026-05-08 07:00:00.000000'),('019f09e1-7e4d-7546-96e7-f586dc4f4e12','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','5','updated','Stats widgets','{}','2026-05-09 08:00:00.000000'),('019f09e1-7e4d-7546-96e7-f587fdecc4d0','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','5','commented','Stats widgets','{}','2026-05-10 09:00:00.000000'),('019f09e1-7e4e-77d9-bbf0-73cb84d3a684','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','6','created','Filter panel UX overhaul','{}','2026-05-07 06:00:00.000000'),('019f09e1-7e4e-77d9-bbf0-73cc3fc8db85','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','6','status_changed','Filter panel UX overhaul','{}','2026-05-08 07:00:00.000000'),('019f09e1-7e4e-77d9-bbf0-73cd6530dbb9','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','6','assigned','Filter panel UX overhaul','{}','2026-05-09 08:00:00.000000'),('019f09e1-7e4f-7007-aec1-3600396ca6d2','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','6','updated','Filter panel UX overhaul','{}','2026-05-10 09:00:00.000000'),('019f09e1-7e4f-7007-aec1-36013e2d054a','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','6','commented','Filter panel UX overhaul','{}','2026-05-11 10:00:00.000000'),('019f09e1-7e4f-7007-aec1-3602540e7f83','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','7','created','Bug: empty-state flicker','{}','2026-05-08 07:00:00.000000'),('019f09e1-7e50-7f62-ac85-c5aaef19877c','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','7','status_changed','Bug: empty-state flicker','{}','2026-05-09 08:00:00.000000'),('019f09e1-7e50-7f62-ac85-c5ab9a4f9142','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','7','assigned','Bug: empty-state flicker','{}','2026-05-10 09:00:00.000000'),('019f09e1-7e50-7f62-ac85-c5ac93b63886','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','7','updated','Bug: empty-state flicker','{}','2026-05-11 10:00:00.000000'),('019f09e1-7e51-7fa6-beba-0c2a5180c2ca','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','7','commented','Bug: empty-state flicker','{}','2026-05-12 11:00:00.000000'),('019f09e1-7e51-7fa6-beba-0c2b552798fa','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','8','created','Dark-mode token migration','{}','2026-05-09 08:00:00.000000'),('019f09e1-7e51-7fa6-beba-0c2c6281a314','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','8','status_changed','Dark-mode token migration','{}','2026-05-10 09:00:00.000000'),('019f09e1-7e52-7d49-b59c-82940a4ea64b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','8','assigned','Dark-mode token migration','{}','2026-05-11 10:00:00.000000'),('019f09e1-7e52-7d49-b59c-8295f6973884','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','8','updated','Dark-mode token migration','{}','2026-05-12 11:00:00.000000'),('019f09e1-7e52-7d49-b59c-8296b0e5be28','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','8','commented','Dark-mode token migration','{}','2026-05-13 12:00:00.000000'),('019f09e1-7e52-7d49-b59c-8297736afba4','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','9','created','API v2 design doc','{}','2026-05-10 09:00:00.000000'),('019f09e1-7e53-7b0c-aa43-924b6ac92956','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','9','status_changed','API v2 design doc','{}','2026-05-11 10:00:00.000000'),('019f09e1-7e53-7b0c-aa43-924c07ec014b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','9','assigned','API v2 design doc','{}','2026-05-12 11:00:00.000000'),('019f09e1-7e54-7603-8e32-b03fd769605d','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','9','updated','API v2 design doc','{}','2026-05-13 12:00:00.000000'),('019f09e1-7e54-7603-8e32-b04074fcee7b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','9','commented','API v2 design doc','{}','2026-05-14 13:00:00.000000'),('019f09e1-7e54-7603-8e32-b04182366bf0','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','10','created','Pagination cursor spec','{}','2026-05-11 10:00:00.000000'),('019f09e1-7e54-7603-8e32-b04232c0bcf7','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','10','status_changed','Pagination cursor spec','{}','2026-05-12 11:00:00.000000'),('019f09e1-7e55-7244-9759-829ec19baa31','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','10','assigned','Pagination cursor spec','{}','2026-05-13 12:00:00.000000'),('019f09e1-7e55-7244-9759-829fcab57a0e','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','10','updated','Pagination cursor spec','{}','2026-05-14 13:00:00.000000'),('019f09e1-7e55-7244-9759-82a0b7556507','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','10','commented','Pagination cursor spec','{}','2026-05-15 14:00:00.000000'),('019f09e1-7e56-7141-85eb-1f224d527d79','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','11','created','OpenAPI 3.1 codegen','{}','2026-05-12 11:00:00.000000'),('019f09e1-7e56-7141-85eb-1f234da23fc4','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','11','status_changed','OpenAPI 3.1 codegen','{}','2026-05-13 12:00:00.000000'),('019f09e1-7e56-7141-85eb-1f2412262c6a','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','11','assigned','OpenAPI 3.1 codegen','{}','2026-05-14 13:00:00.000000'),('019f09e1-7e57-7c86-9f59-2294bfd14287','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','11','updated','OpenAPI 3.1 codegen','{}','2026-05-15 14:00:00.000000'),('019f09e1-7e57-7c86-9f59-2295d1af6beb','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','11','commented','OpenAPI 3.1 codegen','{}','2026-05-16 15:00:00.000000'),('019f09e1-7e57-7c86-9f59-229690a3078a','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','12','created','Rate limiter (sliding window)','{}','2026-05-13 12:00:00.000000'),('019f09e1-7e58-7eb7-a385-95e201e3a050','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','12','status_changed','Rate limiter (sliding window)','{}','2026-05-14 13:00:00.000000'),('019f09e1-7e58-7eb7-a385-95e3257e691a','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','12','assigned','Rate limiter (sliding window)','{}','2026-05-15 14:00:00.000000'),('019f09e1-7e58-7eb7-a385-95e45da77723','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','12','updated','Rate limiter (sliding window)','{}','2026-05-16 15:00:00.000000'),('019f09e1-7e58-7eb7-a385-95e54141db7f','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','12','commented','Rate limiter (sliding window)','{}','2026-05-17 16:00:00.000000'),('019f09e1-7e59-7832-ab05-5dc47aeb4feb','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','13','created','GraphQL N+1 audit','{}','2026-05-14 13:00:00.000000'),('019f09e1-7e59-7832-ab05-5dc5ba39f527','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','13','status_changed','GraphQL N+1 audit','{}','2026-05-15 14:00:00.000000'),('019f09e1-7e59-7832-ab05-5dc6dac44cfb','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','13','assigned','GraphQL N+1 audit','{}','2026-05-16 15:00:00.000000'),('019f09e1-7e5a-791f-b874-d8947e2fafe0','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','13','updated','GraphQL N+1 audit','{}','2026-05-17 16:00:00.000000'),('019f09e1-7e5a-791f-b874-d895a98ab35a','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','13','commented','GraphQL N+1 audit','{}','2026-05-18 17:00:00.000000'),('019f09e1-7e5a-791f-b874-d89628e1d6c1','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','14','created','Bug: 502 on long-polling','{}','2026-05-15 14:00:00.000000'),('019f09e1-7e5b-7a1d-b4a8-9b5ec6b55580','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','14','status_changed','Bug: 502 on long-polling','{}','2026-05-16 15:00:00.000000'),('019f09e1-7e5b-7a1d-b4a8-9b5f0dd58081','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','14','assigned','Bug: 502 on long-polling','{}','2026-05-17 16:00:00.000000'),('019f09e1-7e5b-7a1d-b4a8-9b60e2709856','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','14','updated','Bug: 502 on long-polling','{}','2026-05-18 17:00:00.000000'),('019f09e1-7e5b-7a1d-b4a8-9b6169f2085c','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','14','commented','Bug: 502 on long-polling','{}','2026-05-19 18:00:00.000000'),('019f09e1-7e5c-718a-aef1-e4fb20df284e','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','15','created','Webhook signature v2','{}','2026-05-16 15:00:00.000000'),('019f09e1-7e5c-718a-aef1-e4fc6b543e4b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','15','status_changed','Webhook signature v2','{}','2026-05-17 16:00:00.000000'),('019f09e1-7e5c-718a-aef1-e4fd50d414f2','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','15','assigned','Webhook signature v2','{}','2026-05-18 17:00:00.000000'),('019f09e1-7e5c-718a-aef1-e4febaaf368b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','15','updated','Webhook signature v2','{}','2026-05-19 18:00:00.000000'),('019f09e1-7e5d-7a9d-9084-001fed36d091','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','15','commented','Webhook signature v2','{}','2026-05-20 19:00:00.000000'),('019f09e1-7e5d-7a9d-9084-00201fc44ce7','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','16','created','Drop legacy /v1/sessions','{}','2026-05-17 16:00:00.000000'),('019f09e1-7e5e-7464-bb53-11ef0144927a','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','16','status_changed','Drop legacy /v1/sessions','{}','2026-05-18 17:00:00.000000'),('019f09e1-7e5e-7464-bb53-11f09b36dbf1','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','16','assigned','Drop legacy /v1/sessions','{}','2026-05-19 18:00:00.000000'),('019f09e1-7e5e-7464-bb53-11f12c875cee','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','16','updated','Drop legacy /v1/sessions','{}','2026-05-20 19:00:00.000000'),('019f09e1-7e5e-7464-bb53-11f2804d2406','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','16','commented','Drop legacy /v1/sessions','{}','2026-05-21 20:00:00.000000'),('019f09e1-7e5f-7ef6-b0a0-fcbf17617e4c','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','17','created','iOS app scaffolding','{}','2026-05-18 17:00:00.000000'),('019f09e1-7e5f-7ef6-b0a0-fcc05303b889','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','17','status_changed','iOS app scaffolding','{}','2026-05-19 18:00:00.000000'),('019f09e1-7e5f-7ef6-b0a0-fcc1840cc585','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','17','assigned','iOS app scaffolding','{}','2026-05-20 19:00:00.000000'),('019f09e1-7e60-78fb-a6ab-bb974cc01a44','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','17','updated','iOS app scaffolding','{}','2026-05-21 20:00:00.000000'),('019f09e1-7e60-78fb-a6ab-bb982dba63de','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','17','commented','iOS app scaffolding','{}','2026-05-22 21:00:00.000000'),('019f09e1-7e60-78fb-a6ab-bb99246b6e98','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','18','created','Android app scaffolding','{}','2026-05-19 18:00:00.000000'),('019f09e1-7e60-78fb-a6ab-bb9a8b6d7ef2','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','18','status_changed','Android app scaffolding','{}','2026-05-20 19:00:00.000000'),('019f09e1-7e61-7516-b378-a9e5070e0019','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','18','assigned','Android app scaffolding','{}','2026-05-21 20:00:00.000000'),('019f09e1-7e61-7516-b378-a9e658f3f059','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','18','updated','Android app scaffolding','{}','2026-05-22 21:00:00.000000'),('019f09e1-7e61-7516-b378-a9e7d22d68da','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','18','commented','Android app scaffolding','{}','2026-05-23 22:00:00.000000'),('019f09e1-7e62-7291-b780-ea44fab5f3e1','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','19','created','Push notification plumbing','{}','2026-05-20 19:00:00.000000'),('019f09e1-7e62-7291-b780-ea4591f15deb','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','19','status_changed','Push notification plumbing','{}','2026-05-21 20:00:00.000000'),('019f09e1-7e62-7291-b780-ea463630a2b7','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','19','assigned','Push notification plumbing','{}','2026-05-22 21:00:00.000000'),('019f09e1-7e63-75e2-b710-c4e5637e6b82','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','19','updated','Push notification plumbing','{}','2026-05-23 22:00:00.000000'),('019f09e1-7e63-75e2-b710-c4e6ef8aba60','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','19','commented','Push notification plumbing','{}','2026-05-24 23:00:00.000000'),('019f09e1-7e63-75e2-b710-c4e7e9c82371','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','20','created','Offline-first cache','{}','2026-05-21 20:00:00.000000'),('019f09e1-7e64-7aec-a2a1-bd28e2a9f8a0','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','20','status_changed','Offline-first cache','{}','2026-05-22 21:00:00.000000'),('019f09e1-7e64-7aec-a2a1-bd29e33a8933','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','20','assigned','Offline-first cache','{}','2026-05-23 22:00:00.000000'),('019f09e1-7e64-7aec-a2a1-bd2a67911d09','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','20','updated','Offline-first cache','{}','2026-05-24 23:00:00.000000'),('019f09e1-7e65-7954-8071-df06b5e987cc','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','20','commented','Offline-first cache','{}','2026-05-25 00:00:00.000000'),('019f09e1-7e65-7954-8071-df07c142ef6b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','21','created','Crash reporting integration','{}','2026-05-22 21:00:00.000000'),('019f09e1-7e65-7954-8071-df088603fdf3','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','21','status_changed','Crash reporting integration','{}','2026-05-23 22:00:00.000000'),('019f09e1-7e65-7954-8071-df093fb610d2','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','21','assigned','Crash reporting integration','{}','2026-05-24 23:00:00.000000'),('019f09e1-7e66-7f8c-a89e-b0027cbc6f16','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','21','updated','Crash reporting integration','{}','2026-05-25 00:00:00.000000'),('019f09e1-7e66-7f8c-a89e-b00359426d72','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','21','commented','Crash reporting integration','{}','2026-05-01 01:00:00.000000'),('019f09e1-7e66-7f8c-a89e-b004028b8137','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','22','created','Bug: scroll jank on iOS 17','{}','2026-05-23 22:00:00.000000'),('019f09e1-7e67-724c-b532-fb525f9bdf4c','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','22','status_changed','Bug: scroll jank on iOS 17','{}','2026-05-24 23:00:00.000000'),('019f09e1-7e67-724c-b532-fb534f2a5f23','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','22','assigned','Bug: scroll jank on iOS 17','{}','2026-05-25 00:00:00.000000'),('019f09e1-7e68-7d11-ae08-004e5d5f5c9d','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','22','updated','Bug: scroll jank on iOS 17','{}','2026-05-01 01:00:00.000000'),('019f09e1-7e68-7d11-ae08-004fbed25b57','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','22','commented','Bug: scroll jank on iOS 17','{}','2026-05-02 02:00:00.000000'),('019f09e1-7e68-7d11-ae08-0050ff0fa385','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','task','23','created','Deep-link routing','{}','2026-05-24 23:00:00.000000'),('019f09e1-7e68-7d11-ae08-005120bad2a8','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','23','status_changed','Deep-link routing','{}','2026-05-25 00:00:00.000000'),('019f09e1-7e69-785e-b37a-99973ba227b0','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','23','assigned','Deep-link routing','{}','2026-05-01 01:00:00.000000'),('019f09e1-7e69-785e-b37a-99988c6cd532','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','23','updated','Deep-link routing','{}','2026-05-02 02:00:00.000000'),('019f09e1-7e69-785e-b37a-99992ae2b823','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','23','commented','Deep-link routing','{}','2026-05-03 03:00:00.000000'),('019f09e1-7e6a-7580-b329-3dced146d960','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','task','24','created','GitHub Actions matrix','{}','2026-05-25 00:00:00.000000'),('019f09e1-7e6a-7580-b329-3dcffb478b3b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','24','status_changed','GitHub Actions matrix','{}','2026-05-01 01:00:00.000000'),('019f09e1-7e6a-7580-b329-3dd0314283f2','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','24','assigned','GitHub Actions matrix','{}','2026-05-02 02:00:00.000000'),('019f09e1-7e6a-7580-b329-3dd12bd89fd7','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','24','updated','GitHub Actions matrix','{}','2026-05-03 03:00:00.000000'),('019f09e1-7e6b-7740-a26e-382e53ea6877','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','24','commented','GitHub Actions matrix','{}','2026-05-04 04:00:00.000000'),('019f09e1-7e6b-7740-a26e-382f7da1fee8','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','task','25','created','Docker compose dev env','{}','2026-05-01 01:00:00.000000'),('019f09e1-7e6b-7740-a26e-38300f06cf4e','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','25','status_changed','Docker compose dev env','{}','2026-05-02 02:00:00.000000'),('019f09e1-7e6c-7442-9bf9-650ef4af0aec','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','25','assigned','Docker compose dev env','{}','2026-05-03 03:00:00.000000'),('019f09e1-7e6c-7442-9bf9-650f9e40fca3','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','25','updated','Docker compose dev env','{}','2026-05-04 04:00:00.000000'),('019f09e1-7e6c-7442-9bf9-651062370579','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','25','commented','Docker compose dev env','{}','2026-05-05 05:00:00.000000'),('019f09e1-7e6d-7599-a88d-cf797888c8ad','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','task','26','created','Observability stack (loki)','{}','2026-05-02 02:00:00.000000'),('019f09e1-7e6d-7599-a88d-cf7a3f188d94','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','26','status_changed','Observability stack (loki)','{}','2026-05-03 03:00:00.000000'),('019f09e1-7e6d-7599-a88d-cf7bff1d8397','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','26','assigned','Observability stack (loki)','{}','2026-05-04 04:00:00.000000'),('019f09e1-7e6d-7599-a88d-cf7c97301368','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','26','updated','Observability stack (loki)','{}','2026-05-05 05:00:00.000000'),('019f09e1-7e6e-7ce6-9e32-21e192a94ef4','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','26','commented','Observability stack (loki)','{}','2026-05-06 06:00:00.000000'),('019f09e1-7e6e-7ce6-9e32-21e2a4baa4d1','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfa860248ef','task','27','created','Secrets rotation runbook','{}','2026-05-03 03:00:00.000000'),('019f09e1-7e6e-7ce6-9e32-21e31c4ad254','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','27','status_changed','Secrets rotation runbook','{}','2026-05-04 04:00:00.000000'),('019f09e1-7e6f-7646-a65c-8e129c834932','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','27','assigned','Secrets rotation runbook','{}','2026-05-05 05:00:00.000000'),('019f09e1-7e6f-7646-a65c-8e13b3fda20c','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','27','updated','Secrets rotation runbook','{}','2026-05-06 06:00:00.000000'),('019f09e1-7e70-70b6-a0e2-d80edea8479e','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','27','commented','Secrets rotation runbook','{}','2026-05-07 07:00:00.000000'),('019f09e1-7e70-70b6-a0e2-d80f5557f3e7','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','task','28','created','Postgres → MySQL benchmark','{}','2026-05-04 04:00:00.000000'),('019f09e1-7e70-70b6-a0e2-d8106f7eba98','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','28','status_changed','Postgres → MySQL benchmark','{}','2026-05-05 05:00:00.000000'),('019f09e1-7e71-7003-aa36-7e955104cea5','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','28','assigned','Postgres → MySQL benchmark','{}','2026-05-06 06:00:00.000000'),('019f09e1-7e71-7003-aa36-7e96e4226aca','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','28','updated','Postgres → MySQL benchmark','{}','2026-05-07 07:00:00.000000'),('019f09e1-7e71-7003-aa36-7e9712cea307','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','28','commented','Postgres → MySQL benchmark','{}','2026-05-08 08:00:00.000000'),('019f09e1-7e72-7b86-abc0-72dfdd2e8a36','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','task','29','created','Bug: flaky e2e on retry','{}','2026-05-05 05:00:00.000000'),('019f09e1-7e72-7b86-abc0-72e0c1e08632','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','29','status_changed','Bug: flaky e2e on retry','{}','2026-05-06 06:00:00.000000'),('019f09e1-7e72-7b86-abc0-72e1d2b337ff','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','29','assigned','Bug: flaky e2e on retry','{}','2026-05-07 07:00:00.000000'),('019f09e1-7e72-7b86-abc0-72e25b5bcf3b','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','29','updated','Bug: flaky e2e on retry','{}','2026-05-08 08:00:00.000000'),('019f09e1-7e73-7733-b9aa-3d52c2a576c1','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','29','commented','Bug: flaky e2e on retry','{}','2026-05-09 09:00:00.000000'),('019f09e1-7e73-7733-b9aa-3d5387b2d679','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfc373fc01b','task','30','created','Tame CI cost','{}','2026-05-06 06:00:00.000000'),('019f09e1-7e74-74bb-af45-72522cac9298','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','task','30','status_changed','Tame CI cost','{}','2026-05-07 07:00:00.000000'),('019f09e1-7e74-74bb-af45-7253c935e351','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','task','30','assigned','Tame CI cost','{}','2026-05-08 08:00:00.000000'),('019f09e1-7e74-74bb-af45-7254117207d0','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','task','30','updated','Tame CI cost','{}','2026-05-09 09:00:00.000000'),('019f09e1-7e75-78fc-a34e-fce38eeccbe5','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','task','30','commented','Tame CI cost','{}','2026-05-10 10:00:00.000000');
/*!40000 ALTER TABLE `activity_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `attachments`
--

DROP TABLE IF EXISTS `attachments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `attachments` (
  `id` char(36) NOT NULL,
  `task_id` bigint unsigned DEFAULT NULL COMMENT 'Either task_id or comment_id must be set',
  `comment_id` bigint unsigned DEFAULT NULL COMMENT 'Either task_id or comment_id must be set',
  `uploader_id` char(36) NOT NULL COMMENT 'References users.id',
  `filename` varchar(255) NOT NULL,
  `mime_type` varchar(127) NOT NULL,
  `size_bytes` bigint unsigned NOT NULL,
  `storage_key` varchar(512) NOT NULL,
  `checksum_sha256` char(64) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_attachments_task` (`task_id`),
  KEY `idx_attachments_comment` (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `attachments`
--

LOCK TABLES `attachments` WRITE;
/*!40000 ALTER TABLE `attachments` DISABLE KEYS */;
INSERT INTO `attachments` VALUES ('019f09e1-7e98-701b-9a1c-797e110cd70e',1,NULL,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','shell-mockup.png','image/png',184320,'s3://dira-attachments/019f09e1-7e98-701b-9a1c-797e110cd70e/shell-mockup.png','019f09e17e98701b9a1c797e110cd70e00000000000000000000000000000000','2026-05-12 00:00:00'),('019f09e1-7e99-74c2-b912-2fba6120f0cc',5,NULL,'019f09e1-7dc7-7272-b99a-ebf5bde55420','widget-spec.pdf','application/pdf',524288,'s3://dira-attachments/019f09e1-7e99-74c2-b912-2fba6120f0cc/widget-spec.pdf','019f09e17e9974c2b9122fba6120f0cc00000000000000000000000000000000','2026-05-12 00:00:00'),('019f09e1-7e99-74c2-b912-2fbbe17a90dc',9,NULL,'019f09e1-7dc7-7272-b99a-ebf752763e91','api-v2-rfc.md','text/markdown',18432,'s3://dira-attachments/019f09e1-7e99-74c2-b912-2fbbe17a90dc/api-v2-rfc.md','019f09e17e9974c2b9122fbbe17a90dc00000000000000000000000000000000','2026-05-12 00:00:00'),('019f09e1-7e9a-7431-afdf-976154c12f3c',14,NULL,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','long-poll-trace.har','application/json',2097152,'s3://dira-attachments/019f09e1-7e9a-7431-afdf-976154c12f3c/long-poll-trace.har','019f09e17e9a7431afdf976154c12f3c00000000000000000000000000000000','2026-05-12 00:00:00'),('019f09e1-7e9a-7431-afdf-97621203177b',26,NULL,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','loki-dashboard.json','application/json',65536,'s3://dira-attachments/019f09e1-7e9a-7431-afdf-97621203177b/loki-dashboard.json','019f09e17e9a7431afdf97621203177b00000000000000000000000000000000','2026-05-12 00:00:00'),('019f09e1-7e9a-7431-afdf-976376f5a45f',NULL,8,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','profile.png','image/png',96000,'s3://dira-attachments/019f09e1-7e9a-7431-afdf-976376f5a45f/profile.png','019f09e17e9a7431afdf976376f5a45f00000000000000000000000000000000','2026-05-13 00:00:00'),('019f09e1-7e9b-7d20-afc5-cccfb73e2c13',NULL,13,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','cert-renewal.txt','text/plain',2048,'s3://dira-attachments/019f09e1-7e9b-7d20-afc5-cccfb73e2c13/cert-renewal.txt','019f09e17e9b7d20afc5cccfb73e2c1300000000000000000000000000000000','2026-05-13 00:00:00'),('019f09e1-7e9b-7d20-afc5-ccd09983b9cd',NULL,18,'019f09e1-7dc7-7272-b99a-ebfc373fc01b','pgbouncer.ini','text/plain',4096,'s3://dira-attachments/019f09e1-7e9b-7d20-afc5-ccd09983b9cd/pgbouncer.ini','019f09e17e9b7d20afc5ccd09983b9cd00000000000000000000000000000000','2026-05-13 00:00:00'),('019f09e1-7e9b-7d20-afc5-ccd1f67d3ac6',NULL,19,'019f09e1-7dc7-7272-b99a-ebfda8a94e2c','connection-pool.png','image/png',240000,'s3://dira-attachments/019f09e1-7e9b-7d20-afc5-ccd1f67d3ac6/connection-pool.png','019f09e17e9b7d20afc5ccd1f67d3ac600000000000000000000000000000000','2026-05-13 00:00:00'),('019f09e1-7e9c-7cc4-9b82-4880400658e7',NULL,10,'019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','spec-revision.diff','text/plain',8192,'s3://dira-attachments/019f09e1-7e9c-7cc4-9b82-4880400658e7/spec-revision.diff','019f09e17e9c7cc49b824880400658e700000000000000000000000000000000','2026-05-13 00:00:00');
/*!40000 ALTER TABLE `attachments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Temporary view structure for view `custom_field_options`
--

DROP TABLE IF EXISTS `custom_field_options`;
/*!50001 DROP VIEW IF EXISTS `custom_field_options`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `custom_field_options` AS SELECT 
 1 AS `custom_field_id`,
 1 AS `workspace_id`,
 1 AS `field_key`,
 1 AS `position`,
 1 AS `value`,
 1 AS `label`*/;
SET character_set_client = @saved_cs_client;

--
-- Table structure for table `custom_field_options_cache`
--

DROP TABLE IF EXISTS `custom_field_options_cache`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `custom_field_options_cache` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `custom_field_id` char(36) NOT NULL COMMENT 'References custom_fields.id; populated by triggers on that table',
  `position` int unsigned NOT NULL,
  `value` varchar(255) NOT NULL,
  `label` varchar(255) NOT NULL,
  `cached_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_cache_field_position` (`custom_field_id`,`position`),
  KEY `idx_cache_value` (`value`),
  KEY `idx_cache_label` (`label`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Denormalized cache of custom_fields.options exploded via JSON_TABLE in triggers';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `custom_field_options_cache`
--

LOCK TABLES `custom_field_options_cache` WRITE;
/*!40000 ALTER TABLE `custom_field_options_cache` DISABLE KEYS */;
INSERT INTO `custom_field_options_cache` VALUES (1,'019f09e1-7e9c-7cc4-9b82-48810705d97b',1,'auth','Auth','2026-06-27 16:20:00'),(2,'019f09e1-7e9c-7cc4-9b82-48810705d97b',2,'analytics','Analytics','2026-06-27 16:20:00'),(3,'019f09e1-7e9c-7cc4-9b82-48810705d97b',3,'platform','Platform','2026-06-27 16:20:00'),(4,'019f09e1-7e9c-7cc4-9b82-48810705d97b',4,'mobile','Mobile','2026-06-27 16:20:00');
/*!40000 ALTER TABLE `custom_field_options_cache` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `custom_fields`
--

DROP TABLE IF EXISTS `custom_fields`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `custom_fields` (
  `id` char(36) NOT NULL,
  `workspace_id` char(36) NOT NULL,
  `key` varchar(64) NOT NULL,
  `label` varchar(160) NOT NULL,
  `field_type` enum('text','number','date','select') NOT NULL,
  `options` json NOT NULL COMMENT 'For select fields: array of {value,label}. For others: empty array.',
  `is_required` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_custom_fields` (`workspace_id`,`key`),
  CONSTRAINT `chk_custom_fields_options_schema` CHECK (json_schema_valid(_utf8mb4'{ "type": "array", "items": { "type": "object", "properties": { "value": {"type": "string"}, "label": {"type": "string"} }, "required": ["value", "label"] } }',`options`))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `custom_fields`
--

LOCK TABLES `custom_fields` WRITE;
/*!40000 ALTER TABLE `custom_fields` DISABLE KEYS */;
INSERT INTO `custom_fields` VALUES ('019f09e1-7e9c-7cc4-9b82-48810705d97b','00000001-0000-7000-8000-000000000001','epic','Epic','select','[{\"label\": \"Auth\", \"value\": \"auth\"}, {\"label\": \"Analytics\", \"value\": \"analytics\"}, {\"label\": \"Platform\", \"value\": \"platform\"}, {\"label\": \"Mobile\", \"value\": \"mobile\"}]',0,'2026-04-15 00:00:00'),('019f09e1-7e9d-7b22-a4e0-e00972298915','00000001-0000-7000-8000-000000000001','story_points','Story Points','number','[]',0,'2026-04-15 00:00:00'),('019f09e1-7e9d-7b22-a4e0-e00a6ed88afc','00000001-0000-7000-8000-000000000001','release','Target Release','text','[]',0,'2026-04-15 00:00:00'),('019f09e1-7e9e-7781-aaee-26da492caefc','00000001-0000-7000-8000-000000000002','customer','Customer','text','[]',0,'2026-04-15 00:00:00');
/*!40000 ALTER TABLE `custom_fields` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'IGNORE_SPACE,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_custom_fields_options_insert` AFTER INSERT ON `custom_fields` FOR EACH ROW BEGIN
            IF NEW.field_type = 'select' AND JSON_LENGTH(NEW.options) > 0 THEN
                INSERT INTO custom_field_options_cache
                    (custom_field_id, position, value, label)
                SELECT NEW.id, opt.position, opt.value, opt.label
                FROM JSON_TABLE(
                    NEW.options,
                    '$[*]' COLUMNS (
                        position FOR ORDINALITY,
                        value VARCHAR(255) PATH '$.value',
                        label VARCHAR(255) PATH '$.label'
                    )
                ) AS opt;
            END IF;
        END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'IGNORE_SPACE,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_custom_fields_options_update` AFTER UPDATE ON `custom_fields` FOR EACH ROW BEGIN
            IF NOT (NEW.options <=> OLD.options)
                OR NOT (NEW.field_type <=> OLD.field_type) THEN
                DELETE FROM custom_field_options_cache
                    WHERE custom_field_id = OLD.id;
                IF NEW.field_type = 'select' AND JSON_LENGTH(NEW.options) > 0 THEN
                    INSERT INTO custom_field_options_cache
                        (custom_field_id, position, value, label)
                    SELECT NEW.id, opt.position, opt.value, opt.label
                    FROM JSON_TABLE(
                        NEW.options,
                        '$[*]' COLUMNS (
                            position FOR ORDINALITY,
                            value VARCHAR(255) PATH '$.value',
                            label VARCHAR(255) PATH '$.label'
                        )
                    ) AS opt;
                END IF;
            END IF;
        END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'IGNORE_SPACE,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_custom_fields_options_delete` AFTER DELETE ON `custom_fields` FOR EACH ROW BEGIN
            DELETE FROM custom_field_options_cache
                WHERE custom_field_id = OLD.id;
        END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `invitations`
--

DROP TABLE IF EXISTS `invitations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `invitations` (
  `id` char(36) NOT NULL,
  `workspace_id` char(36) NOT NULL,
  `inviter_id` char(36) NOT NULL COMMENT 'References users.id',
  `email` varchar(255) NOT NULL,
  `role` enum('admin','member','guest') NOT NULL DEFAULT 'member',
  `accepted_user_id` char(36) DEFAULT NULL COMMENT 'Set when invitation is accepted; references users.id',
  `accepted_at` datetime DEFAULT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_invitations_workspace` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invitations`
--

LOCK TABLES `invitations` WRITE;
/*!40000 ALTER TABLE `invitations` DISABLE KEYS */;
INSERT INTO `invitations` VALUES ('019f09e1-7e0f-7944-9bbd-fa4b49455ac6','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','newhire@dira.test','member',NULL,NULL,'2026-06-30 00:00:00','2026-05-20 00:00:00'),('019f09e1-7e0f-7944-9bbd-fa4cab4e29cc','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','contractor@external.test','guest',NULL,NULL,'2026-07-01 00:00:00','2026-05-22 00:00:00'),('019f09e1-7e0f-7944-9bbd-fa4d90b443ee','00000001-0000-7000-8000-000000000002','019f09e1-7dc7-7272-b99a-ebfa860248ef','priya@dira.test','member','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','2026-04-10 00:00:00','2026-05-01 00:00:00','2026-04-01 00:00:00');
/*!40000 ALTER TABLE `invitations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `kysely_migration`
--

DROP TABLE IF EXISTS `kysely_migration`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `kysely_migration` (
  `name` varchar(255) NOT NULL,
  `timestamp` varchar(255) NOT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `kysely_migration`
--

LOCK TABLES `kysely_migration` WRITE;
/*!40000 ALTER TABLE `kysely_migration` DISABLE KEYS */;
INSERT INTO `kysely_migration` VALUES ('s1_001_users','2026-06-27T16:19:59.546Z'),('s1_002_projects','2026-06-27T16:19:59.552Z'),('s1_003_tasks_and_comments','2026-06-27T16:19:59.561Z'),('s1_004_labels_and_watchers','2026-06-27T16:19:59.569Z'),('s2_001_workspaces_and_membership','2026-06-27T16:19:59.590Z'),('s3_001_sprints','2026-06-27T16:19:59.596Z'),('s3_002_multi_assignees','2026-06-27T16:19:59.599Z'),('s3_003_activity_feed','2026-06-27T16:19:59.606Z'),('s4_001_time_tracking','2026-06-27T16:19:59.616Z'),('s4_002_attachments','2026-06-27T16:19:59.621Z'),('s4_003_custom_fields','2026-06-27T16:19:59.627Z'),('s4_004_custom_field_options_view','2026-06-27T16:19:59.628Z'),('s4_005_custom_field_options_cache','2026-06-27T16:19:59.635Z');
/*!40000 ALTER TABLE `kysely_migration` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `kysely_migration_lock`
--

DROP TABLE IF EXISTS `kysely_migration_lock`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `kysely_migration_lock` (
  `id` varchar(255) NOT NULL,
  `is_locked` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `kysely_migration_lock`
--

LOCK TABLES `kysely_migration_lock` WRITE;
/*!40000 ALTER TABLE `kysely_migration_lock` DISABLE KEYS */;
INSERT INTO `kysely_migration_lock` VALUES ('migration_lock',0);
/*!40000 ALTER TABLE `kysely_migration_lock` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `labels`
--

DROP TABLE IF EXISTS `labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `labels` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` char(36) NOT NULL,
  `name` varchar(64) NOT NULL,
  `color` char(7) NOT NULL DEFAULT '#888888' COMMENT 'Hex color',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_labels_project_name` (`project_id`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `labels`
--

LOCK TABLES `labels` WRITE;
/*!40000 ALTER TABLE `labels` DISABLE KEYS */;
INSERT INTO `labels` VALUES (1,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','bug','#d73a4a','2026-03-27 00:00:00'),(2,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','feature','#0075ca','2026-03-27 00:00:00'),(3,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','docs','#0052cc','2026-03-27 00:00:00'),(4,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','refactor','#5319e7','2026-03-27 00:00:00'),(5,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','quick-win','#7057ff','2026-03-27 00:00:00'),(6,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','tech-debt','#fbca04','2026-03-27 00:00:00'),(7,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','bug','#d73a4a','2026-03-27 00:00:00'),(8,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','feature','#0075ca','2026-03-27 00:00:00'),(9,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','docs','#0052cc','2026-03-27 00:00:00'),(10,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','refactor','#5319e7','2026-03-27 00:00:00'),(11,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','quick-win','#7057ff','2026-03-27 00:00:00'),(12,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','tech-debt','#fbca04','2026-03-27 00:00:00'),(13,'019f09e1-7dc7-7272-b99a-ec004c2014e5','bug','#d73a4a','2026-03-27 00:00:00'),(14,'019f09e1-7dc7-7272-b99a-ec004c2014e5','feature','#0075ca','2026-03-27 00:00:00'),(15,'019f09e1-7dc7-7272-b99a-ec004c2014e5','docs','#0052cc','2026-03-27 00:00:00'),(16,'019f09e1-7dc7-7272-b99a-ec004c2014e5','refactor','#5319e7','2026-03-27 00:00:00'),(17,'019f09e1-7dc7-7272-b99a-ec004c2014e5','quick-win','#7057ff','2026-03-27 00:00:00'),(18,'019f09e1-7dc7-7272-b99a-ec004c2014e5','tech-debt','#fbca04','2026-03-27 00:00:00'),(19,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','bug','#d73a4a','2026-03-27 00:00:00'),(20,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','feature','#0075ca','2026-03-27 00:00:00'),(21,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','docs','#0052cc','2026-03-27 00:00:00'),(22,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','refactor','#5319e7','2026-03-27 00:00:00'),(23,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','quick-win','#7057ff','2026-03-27 00:00:00'),(24,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','tech-debt','#fbca04','2026-03-27 00:00:00');
/*!40000 ALTER TABLE `labels` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `mentions`
--

DROP TABLE IF EXISTS `mentions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `mentions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `comment_id` bigint unsigned NOT NULL,
  `mentioned_user_id` char(36) NOT NULL COMMENT 'References users.id',
  `notified_at` datetime DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_mentions` (`comment_id`,`mentioned_user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mentions`
--

LOCK TABLES `mentions` WRITE;
/*!40000 ALTER TABLE `mentions` DISABLE KEYS */;
INSERT INTO `mentions` VALUES (1,1,'019f09e1-7dc7-7272-b99a-ebfa860248ef','2026-05-15 00:00:00','2026-05-15 00:00:00'),(2,3,'019f09e1-7dc7-7272-b99a-ebf5bde55420','2026-05-15 00:00:00','2026-05-15 00:00:00'),(3,5,'019f09e1-7dc7-7272-b99a-ebfda8a94e2c','2026-05-15 00:00:00','2026-05-15 00:00:00'),(4,7,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-05-15 00:00:00','2026-05-15 00:00:00'),(5,9,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','2026-05-15 00:00:00','2026-05-15 00:00:00'),(6,11,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-15 00:00:00','2026-05-15 00:00:00'),(7,13,'019f09e1-7dc7-7272-b99a-ebfa860248ef','2026-05-15 00:00:00','2026-05-15 00:00:00'),(8,15,'019f09e1-7dc7-7272-b99a-ebf5bde55420','2026-05-15 00:00:00','2026-05-15 00:00:00'),(9,17,'019f09e1-7dc7-7272-b99a-ebfda8a94e2c','2026-05-15 00:00:00','2026-05-15 00:00:00');
/*!40000 ALTER TABLE `mentions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `projects`
--

DROP TABLE IF EXISTS `projects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `projects` (
  `id` char(36) NOT NULL,
  `workspace_id` char(36) DEFAULT NULL COMMENT 'Added stage 2; nullable so dbctx must still infer the join',
  `owner_id` char(36) NOT NULL COMMENT 'References users.id',
  `key` varchar(16) NOT NULL COMMENT 'Short code like WEB, API, MOB',
  `name` varchar(160) NOT NULL,
  `description` mediumtext,
  `visibility` enum('private','workspace','public') NOT NULL DEFAULT 'workspace',
  `is_archived` tinyint(1) NOT NULL DEFAULT '0',
  `settings` json NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_projects_key` (`key`),
  KEY `idx_projects_workspace` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `projects`
--

LOCK TABLES `projects` WRITE;
/*!40000 ALTER TABLE `projects` DISABLE KEYS */;
INSERT INTO `projects` VALUES ('019f09e1-7dc7-7272-b99a-ebfe9e3ff956','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','WEB','Web Dashboard','Customer-facing analytics dashboard.','workspace',0,'{}','2026-02-15 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebfffbda1e2e','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','API','Public API','External REST/GraphQL surface.','workspace',0,'{}','2026-02-15 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ec004c2014e5','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','MOB','Mobile Apps','iOS + Android client apps.','workspace',0,'{}','2026-02-15 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ec01b3ac2367','00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','INF','Infra & Platform','CI/CD, observability, dev environments.','private',0,'{}','2026-02-15 00:00:00','2026-06-27 16:20:00');
/*!40000 ALTER TABLE `projects` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sprints`
--

DROP TABLE IF EXISTS `sprints`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sprints` (
  `id` char(36) NOT NULL,
  `project_id` char(36) NOT NULL,
  `name` varchar(160) NOT NULL,
  `goal` mediumtext,
  `state` enum('planned','active','completed') NOT NULL DEFAULT 'planned',
  `starts_at` datetime NOT NULL,
  `ends_at` datetime NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sprints_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sprints`
--

LOCK TABLES `sprints` WRITE;
/*!40000 ALTER TABLE `sprints` DISABLE KEYS */;
INSERT INTO `sprints` VALUES ('019f09e1-7e2d-7acf-81a2-05cd5053fa8a','019f09e1-7dc7-7272-b99a-ebfe9e3ff956','WEB Sprint 1','Dashboard MVP','completed','2026-04-01 00:00:00','2026-04-14 23:59:59','2026-04-01 00:00:00','2026-04-14 23:59:59'),('019f09e1-7e2e-766b-97df-c0761febb730','019f09e1-7dc7-7272-b99a-ebfe9e3ff956','WEB Sprint 2','Auth + widgets','active','2026-05-15 00:00:00','2026-05-29 23:59:59','2026-05-15 00:00:00','2026-05-29 23:59:59'),('019f09e1-7e2e-766b-97df-c077054be3ca','019f09e1-7dc7-7272-b99a-ebfffbda1e2e','API Sprint 1','v2 cursor pagination','active','2026-05-12 00:00:00','2026-05-26 23:59:59','2026-05-12 00:00:00','2026-05-26 23:59:59'),('019f09e1-7e2e-766b-97df-c0786475a821','019f09e1-7dc7-7272-b99a-ebfffbda1e2e','API Sprint 2','Rate limit + webhooks','planned','2026-05-27 00:00:00','2026-06-10 23:59:59','2026-05-27 00:00:00','2026-06-10 23:59:59'),('019f09e1-7e2f-7970-bc23-72a9e3a61314','019f09e1-7dc7-7272-b99a-ec004c2014e5','MOB Sprint 1','Push + crash reporting','active','2026-05-10 00:00:00','2026-05-24 23:59:59','2026-05-10 00:00:00','2026-05-24 23:59:59'),('019f09e1-7e30-7694-a938-b92215ee9073','019f09e1-7dc7-7272-b99a-ec01b3ac2367','INF Sprint 1','Observability','active','2026-05-13 00:00:00','2026-05-27 23:59:59','2026-05-13 00:00:00','2026-05-27 23:59:59');
/*!40000 ALTER TABLE `sprints` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_assignees`
--

DROP TABLE IF EXISTS `task_assignees`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `task_assignees` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `user_id` char(36) NOT NULL,
  `assigned_by_id` char(36) DEFAULT NULL COMMENT 'References users.id; null for system assignments',
  `assigned_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_task_assignees` (`task_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=35 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_assignees`
--

LOCK TABLES `task_assignees` WRITE;
/*!40000 ALTER TABLE `task_assignees` DISABLE KEYS */;
INSERT INTO `task_assignees` VALUES (1,1,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(2,1,'019f09e1-7dc7-7272-b99a-ebf5bde55420','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-11 00:00:00'),(3,2,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(4,3,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(5,4,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(6,4,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-11 00:00:00'),(7,5,'019f09e1-7dc7-7272-b99a-ebf5bde55420','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(8,6,'019f09e1-7dc7-7272-b99a-ebf5bde55420','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(9,7,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(10,7,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-11 00:00:00'),(11,9,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(12,10,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(13,11,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(14,12,'019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(15,13,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(16,14,'019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(17,14,'019f09e1-7dc7-7272-b99a-ebf5bde55420','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-11 00:00:00'),(18,16,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(19,17,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(20,18,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(21,18,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-11 00:00:00'),(22,19,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(23,20,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(24,21,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(25,22,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(26,24,'019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(27,25,'019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(28,25,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-11 00:00:00'),(29,26,'019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(30,27,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(31,28,'019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(32,28,'019f09e1-7dc7-7272-b99a-ebf5bde55420','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-11 00:00:00'),(33,29,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00'),(34,30,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-10 00:00:00');
/*!40000 ALTER TABLE `task_assignees` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_comments`
--

DROP TABLE IF EXISTS `task_comments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `task_comments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `author_id` char(36) NOT NULL COMMENT 'References users.id',
  `parent_comment_id` bigint unsigned DEFAULT NULL COMMENT 'Self-ref for threaded replies',
  `body` mediumtext NOT NULL,
  `is_edited` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_comments_task` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_comments`
--

LOCK TABLES `task_comments` WRITE;
/*!40000 ALTER TABLE `task_comments` DISABLE KEYS */;
INSERT INTO `task_comments` VALUES (1,1,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f',NULL,'Shell PR is up — going through review.',0,'2026-04-16 00:00:00','2026-04-17 00:00:00'),(2,1,'019f09e1-7dc7-7272-b99a-ebf2ed933f35',1,'Merged. Thanks!',0,'2026-04-17 00:00:00','2026-04-18 00:00:00'),(3,4,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f',NULL,'Rotation interval — 30 days?',0,'2026-04-18 00:00:00','2026-04-19 00:00:00'),(4,4,'019f09e1-7dc7-7272-b99a-ebf2ed933f35',3,'Lets start with 14 and revisit.',0,'2026-04-19 00:00:00','2026-04-20 00:00:00'),(5,4,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f',4,'14 it is.',0,'2026-04-20 00:00:00','2026-04-21 00:00:00'),(6,6,'019f09e1-7dc7-7272-b99a-ebf5bde55420',NULL,'Mock done, hooking up filters now.',0,'2026-04-21 00:00:00','2026-04-22 00:00:00'),(7,7,'019f09e1-7dc7-7272-b99a-ebf62a8f2178',NULL,'Reproduces only on cold-start.',0,'2026-04-22 00:00:00','2026-04-23 00:00:00'),(8,7,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f',7,'Likely the suspense boundary.',0,'2026-04-23 00:00:00','2026-04-24 00:00:00'),(9,11,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f',NULL,'Spec covers ULID cursors. Comments?',0,'2026-04-24 00:00:00','2026-04-25 00:00:00'),(10,11,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',9,'Looks good. Ship it.',0,'2026-04-25 00:00:00','2026-04-26 00:00:00'),(11,14,'019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'Identified upstream — nginx idle timeout.',0,'2026-04-26 00:00:00','2026-04-27 00:00:00'),(12,14,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',11,'Bumping to 75s should be enough.',0,'2026-04-27 00:00:00','2026-04-28 00:00:00'),(13,19,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d',NULL,'APNs cert renewed.',0,'2026-04-28 00:00:00','2026-04-29 00:00:00'),(14,19,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',13,'FCM token bridge next.',0,'2026-04-29 00:00:00','2026-04-30 00:00:00'),(15,21,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d',NULL,'Sentry SDK pinned.',0,'2026-04-30 00:00:00','2026-05-01 00:00:00'),(16,26,'019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'Loki + promtail wired. Grafana dashboards WIP.',0,'2026-05-01 00:00:00','2026-05-02 00:00:00'),(17,26,'019f09e1-7dc7-7272-b99a-ebf2ed933f35',16,'Add the SLO board too.',0,'2026-05-02 00:00:00','2026-05-03 00:00:00'),(18,29,'019f09e1-7dc7-7272-b99a-ebf62a8f2178',NULL,'Repro: postgres pool exhaustion under burst.',0,'2026-05-03 00:00:00','2026-05-04 00:00:00'),(19,29,'019f09e1-7dc7-7272-b99a-ebf752763e91',18,'Bumping max connections + adding pgbouncer.',0,'2026-05-04 00:00:00','2026-05-05 00:00:00'),(20,5,'019f09e1-7dc7-7272-b99a-ebf5bde55420',NULL,'Bar charts done; line charts next.',0,'2026-05-05 00:00:00','2026-05-06 00:00:00');
/*!40000 ALTER TABLE `task_comments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_custom_field_values`
--

DROP TABLE IF EXISTS `task_custom_field_values`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `task_custom_field_values` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `custom_field_id` char(36) NOT NULL,
  `value_text` varchar(1024) DEFAULT NULL,
  `value_number` decimal(18,4) DEFAULT NULL,
  `value_date` date DEFAULT NULL,
  `selected_option` varchar(255) DEFAULT NULL,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_task_custom_field_values` (`task_id`,`custom_field_id`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_custom_field_values`
--

LOCK TABLES `task_custom_field_values` WRITE;
/*!40000 ALTER TABLE `task_custom_field_values` DISABLE KEYS */;
INSERT INTO `task_custom_field_values` VALUES (1,1,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'auth','2026-05-20 00:00:00'),(2,1,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,2.0000,NULL,NULL,'2026-05-20 00:00:00'),(3,1,'019f09e1-7e9d-7b22-a4e0-e00a6ed88afc','2026.5.0',NULL,NULL,NULL,'2026-05-20 00:00:00'),(4,4,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'platform','2026-05-20 00:00:00'),(5,5,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,10.0000,NULL,NULL,'2026-05-20 00:00:00'),(6,6,'019f09e1-7e9d-7b22-a4e0-e00a6ed88afc','2026.7.0',NULL,NULL,NULL,'2026-05-20 00:00:00'),(7,7,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'auth','2026-05-20 00:00:00'),(8,9,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,8.0000,NULL,NULL,'2026-05-20 00:00:00'),(9,10,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'mobile','2026-05-20 00:00:00'),(10,11,'019f09e1-7e9d-7b22-a4e0-e00a6ed88afc','2026.6.0',NULL,NULL,NULL,'2026-05-20 00:00:00'),(11,13,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'analytics','2026-05-20 00:00:00'),(12,13,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,6.0000,NULL,NULL,'2026-05-20 00:00:00'),(13,16,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'auth','2026-05-20 00:00:00'),(14,16,'019f09e1-7e9d-7b22-a4e0-e00a6ed88afc','2026.5.0',NULL,NULL,NULL,'2026-05-20 00:00:00'),(15,17,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,4.0000,NULL,NULL,'2026-05-20 00:00:00'),(16,19,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'platform','2026-05-20 00:00:00'),(17,21,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,2.0000,NULL,NULL,'2026-05-20 00:00:00'),(18,21,'019f09e1-7e9d-7b22-a4e0-e00a6ed88afc','2026.7.0',NULL,NULL,NULL,'2026-05-20 00:00:00'),(19,22,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'auth','2026-05-20 00:00:00'),(20,25,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'mobile','2026-05-20 00:00:00'),(21,25,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,10.0000,NULL,NULL,'2026-05-20 00:00:00'),(22,26,'019f09e1-7e9d-7b22-a4e0-e00a6ed88afc','2026.6.0',NULL,NULL,NULL,'2026-05-20 00:00:00'),(23,28,'019f09e1-7e9c-7cc4-9b82-48810705d97b',NULL,NULL,NULL,'analytics','2026-05-20 00:00:00'),(24,29,'019f09e1-7e9d-7b22-a4e0-e00972298915',NULL,8.0000,NULL,NULL,'2026-05-20 00:00:00');
/*!40000 ALTER TABLE `task_custom_field_values` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_labels`
--

DROP TABLE IF EXISTS `task_labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `task_labels` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `label_id` bigint unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_task_labels` (`task_id`,`label_id`)
) ENGINE=InnoDB AUTO_INCREMENT=33 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_labels`
--

LOCK TABLES `task_labels` WRITE;
/*!40000 ALTER TABLE `task_labels` DISABLE KEYS */;
INSERT INTO `task_labels` VALUES (1,1,2,'2026-03-17 00:00:00'),(2,2,2,'2026-03-18 00:00:00'),(3,3,2,'2026-03-19 00:00:00'),(4,4,2,'2026-03-20 00:00:00'),(5,4,6,'2026-03-20 00:00:00'),(6,5,2,'2026-03-21 00:00:00'),(7,6,2,'2026-03-22 00:00:00'),(8,7,1,'2026-03-23 00:00:00'),(9,8,4,'2026-03-24 00:00:00'),(10,9,9,'2026-03-25 00:00:00'),(11,10,8,'2026-03-26 00:00:00'),(12,11,8,'2026-03-27 00:00:00'),(13,12,8,'2026-03-28 00:00:00'),(14,13,12,'2026-03-29 00:00:00'),(15,14,7,'2026-03-30 00:00:00'),(16,15,8,'2026-03-31 00:00:00'),(17,16,12,'2026-04-01 00:00:00'),(18,17,14,'2026-04-02 00:00:00'),(19,18,14,'2026-04-03 00:00:00'),(20,19,14,'2026-04-04 00:00:00'),(21,20,14,'2026-04-05 00:00:00'),(22,20,18,'2026-04-05 00:00:00'),(23,21,14,'2026-04-06 00:00:00'),(24,22,13,'2026-04-07 00:00:00'),(25,23,14,'2026-04-08 00:00:00'),(26,24,20,'2026-04-09 00:00:00'),(27,25,20,'2026-04-10 00:00:00'),(28,26,20,'2026-04-11 00:00:00'),(29,27,21,'2026-04-12 00:00:00'),(30,28,24,'2026-04-13 00:00:00'),(31,29,19,'2026-04-14 00:00:00'),(32,30,24,'2026-04-15 00:00:00');
/*!40000 ALTER TABLE `task_labels` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_sprints`
--

DROP TABLE IF EXISTS `task_sprints`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `task_sprints` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `sprint_id` char(36) NOT NULL,
  `added_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `removed_at` datetime DEFAULT NULL COMMENT 'Tasks can move between sprints; null means currently in this sprint',
  PRIMARY KEY (`id`),
  KEY `idx_task_sprints_task` (`task_id`),
  KEY `idx_task_sprints_sprint` (`sprint_id`)
) ENGINE=InnoDB AUTO_INCREMENT=37 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_sprints`
--

LOCK TABLES `task_sprints` WRITE;
/*!40000 ALTER TABLE `task_sprints` DISABLE KEYS */;
INSERT INTO `task_sprints` VALUES (1,1,'019f09e1-7e2d-7acf-81a2-05cd5053fa8a','2026-04-01 00:00:00','2026-04-14 23:59:59'),(2,2,'019f09e1-7e2d-7acf-81a2-05cd5053fa8a','2026-04-01 00:00:00','2026-04-14 23:59:59'),(3,3,'019f09e1-7e2d-7acf-81a2-05cd5053fa8a','2026-04-01 00:00:00','2026-04-14 23:59:59'),(4,4,'019f09e1-7e2d-7acf-81a2-05cd5053fa8a','2026-04-01 00:00:00','2026-04-14 23:59:59'),(5,5,'019f09e1-7e2d-7acf-81a2-05cd5053fa8a','2026-04-01 00:00:00','2026-04-14 23:59:59'),(6,6,'019f09e1-7e2d-7acf-81a2-05cd5053fa8a','2026-04-01 00:00:00','2026-04-14 23:59:59'),(7,1,'019f09e1-7e2e-766b-97df-c0761febb730','2026-05-15 00:00:00',NULL),(8,2,'019f09e1-7e2e-766b-97df-c0761febb730','2026-05-15 00:00:00',NULL),(9,3,'019f09e1-7e2e-766b-97df-c0761febb730','2026-05-15 00:00:00',NULL),(10,4,'019f09e1-7e2e-766b-97df-c0761febb730','2026-05-15 00:00:00',NULL),(11,5,'019f09e1-7e2e-766b-97df-c0761febb730','2026-05-15 00:00:00',NULL),(12,6,'019f09e1-7e2e-766b-97df-c0761febb730','2026-05-15 00:00:00',NULL),(13,9,'019f09e1-7e2e-766b-97df-c077054be3ca','2026-05-12 00:00:00',NULL),(14,10,'019f09e1-7e2e-766b-97df-c077054be3ca','2026-05-12 00:00:00',NULL),(15,11,'019f09e1-7e2e-766b-97df-c077054be3ca','2026-05-12 00:00:00',NULL),(16,12,'019f09e1-7e2e-766b-97df-c077054be3ca','2026-05-12 00:00:00',NULL),(17,13,'019f09e1-7e2e-766b-97df-c077054be3ca','2026-05-12 00:00:00',NULL),(18,14,'019f09e1-7e2e-766b-97df-c077054be3ca','2026-05-12 00:00:00',NULL),(19,9,'019f09e1-7e2e-766b-97df-c0786475a821','2026-05-27 00:00:00',NULL),(20,10,'019f09e1-7e2e-766b-97df-c0786475a821','2026-05-27 00:00:00',NULL),(21,11,'019f09e1-7e2e-766b-97df-c0786475a821','2026-05-27 00:00:00',NULL),(22,12,'019f09e1-7e2e-766b-97df-c0786475a821','2026-05-27 00:00:00',NULL),(23,13,'019f09e1-7e2e-766b-97df-c0786475a821','2026-05-27 00:00:00',NULL),(24,14,'019f09e1-7e2e-766b-97df-c0786475a821','2026-05-27 00:00:00',NULL),(25,17,'019f09e1-7e2f-7970-bc23-72a9e3a61314','2026-05-10 00:00:00',NULL),(26,18,'019f09e1-7e2f-7970-bc23-72a9e3a61314','2026-05-10 00:00:00',NULL),(27,19,'019f09e1-7e2f-7970-bc23-72a9e3a61314','2026-05-10 00:00:00',NULL),(28,20,'019f09e1-7e2f-7970-bc23-72a9e3a61314','2026-05-10 00:00:00',NULL),(29,21,'019f09e1-7e2f-7970-bc23-72a9e3a61314','2026-05-10 00:00:00',NULL),(30,22,'019f09e1-7e2f-7970-bc23-72a9e3a61314','2026-05-10 00:00:00',NULL),(31,24,'019f09e1-7e30-7694-a938-b92215ee9073','2026-05-13 00:00:00',NULL),(32,25,'019f09e1-7e30-7694-a938-b92215ee9073','2026-05-13 00:00:00',NULL),(33,26,'019f09e1-7e30-7694-a938-b92215ee9073','2026-05-13 00:00:00',NULL),(34,27,'019f09e1-7e30-7694-a938-b92215ee9073','2026-05-13 00:00:00',NULL),(35,28,'019f09e1-7e30-7694-a938-b92215ee9073','2026-05-13 00:00:00',NULL),(36,29,'019f09e1-7e30-7694-a938-b92215ee9073','2026-05-13 00:00:00',NULL);
/*!40000 ALTER TABLE `task_sprints` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `task_watchers`
--

DROP TABLE IF EXISTS `task_watchers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `task_watchers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `user_id` char(36) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_task_watchers` (`task_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `task_watchers`
--

LOCK TABLES `task_watchers` WRITE;
/*!40000 ALTER TABLE `task_watchers` DISABLE KEYS */;
INSERT INTO `task_watchers` VALUES (1,1,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-03-27 00:00:00'),(2,1,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-03-28 00:00:00'),(3,1,'019f09e1-7dc7-7272-b99a-ebf5bde55420','2026-03-29 00:00:00'),(4,4,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-03-30 00:00:00'),(5,4,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-03-31 00:00:00'),(6,7,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-04-01 00:00:00'),(7,7,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','2026-04-02 00:00:00'),(8,11,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-04-03 00:00:00'),(9,11,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-04-04 00:00:00'),(10,14,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-04-05 00:00:00'),(11,14,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-04-06 00:00:00'),(12,14,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-04-07 00:00:00'),(13,19,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','2026-04-08 00:00:00'),(14,19,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-04-09 00:00:00'),(15,26,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-04-10 00:00:00'),(16,26,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-04-11 00:00:00'),(17,29,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','2026-04-12 00:00:00'),(18,29,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-04-13 00:00:00');
/*!40000 ALTER TABLE `task_watchers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tasks`
--

DROP TABLE IF EXISTS `tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` char(36) NOT NULL,
  `reporter_id` char(36) NOT NULL COMMENT 'References users.id',
  `assignee_id` char(36) DEFAULT NULL COMMENT 'Primary assignee. After stage 3 this becomes a denormalized cache alongside task_assignees',
  `parent_task_id` bigint unsigned DEFAULT NULL COMMENT 'Self-referencing FK by naming',
  `title` varchar(255) NOT NULL,
  `description` mediumtext,
  `status` enum('todo','in_progress','in_review','done','cancelled') NOT NULL DEFAULT 'todo',
  `priority` enum('low','medium','high','urgent') NOT NULL DEFAULT 'medium',
  `is_archived` tinyint(1) NOT NULL DEFAULT '0',
  `due_date` datetime DEFAULT NULL,
  `position` int NOT NULL DEFAULT '0',
  `estimate_hours` decimal(6,2) DEFAULT NULL COMMENT 'Added stage 4; nullable estimate in hours',
  `metadata` json NOT NULL,
  `title_lower` varchar(255) GENERATED ALWAYS AS (lower(`title`)) VIRTUAL COMMENT 'VIRTUAL generated column — indexed for case-insensitive search',
  `title_length` int unsigned GENERATED ALWAYS AS (char_length(`title`)) STORED COMMENT 'STORED generated column — materialized for cheap ORDER BY',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tasks_project` (`project_id`),
  KEY `idx_tasks_status` (`status`),
  KEY `idx_tasks_title_length` (`title_length`),
  KEY `idx_tasks_title_lower` (`title_lower`),
  CONSTRAINT `chk_tasks_position_non_negative` CHECK ((`position` >= 0))
) ENGINE=InnoDB AUTO_INCREMENT=31 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tasks`
--

LOCK TABLES `tasks` WRITE;
/*!40000 ALTER TABLE `tasks` DISABLE KEYS */;
INSERT INTO `tasks` (`id`, `project_id`, `reporter_id`, `assignee_id`, `parent_task_id`, `title`, `description`, `status`, `priority`, `is_archived`, `due_date`, `position`, `estimate_hours`, `metadata`, `created_at`, `updated_at`) VALUES (1,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf2ed933f35',NULL,'Dashboard MVP shell','Detail for Dashboard MVP shell.','done','high',0,'2026-04-26 00:00:00',0,3.50,'{}','2026-03-07 00:00:00','2026-06-27 16:20:00'),(2,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf3fd700a5f',NULL,'Token-based auth','Detail for Token-based auth.','done','high',0,'2026-05-01 00:00:00',10,5.50,'{}','2026-03-08 00:00:00','2026-06-27 16:20:00'),(3,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebf5bde55420','019f09e1-7dc7-7272-b99a-ebf3fd700a5f',2,'Hook up Clerk session','Detail for Hook up Clerk session.','done','high',0,'2026-05-04 00:00:00',20,7.50,'{}','2026-03-09 00:00:00','2026-06-27 16:20:00'),(4,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebf62a8f2178','019f09e1-7dc7-7272-b99a-ebf3fd700a5f',2,'Refresh-token rotation','Detail for Refresh-token rotation.','in_review','high',0,'2026-05-21 00:00:00',30,9.50,'{}','2026-03-10 00:00:00','2026-06-27 16:20:00'),(5,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf5bde55420',NULL,'Stats widgets','Detail for Stats widgets.','in_progress','medium',0,'2026-05-30 00:00:00',40,1.50,'{}','2026-03-11 00:00:00','2026-06-27 16:20:00'),(6,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','019f09e1-7dc7-7272-b99a-ebf5bde55420',NULL,'Filter panel UX overhaul','Detail for Filter panel UX overhaul.','todo','medium',0,'2026-06-13 00:00:00',50,3.50,'{}','2026-03-12 00:00:00','2026-06-27 16:20:00'),(7,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','019f09e1-7dc7-7272-b99a-ebf62a8f2178',NULL,'Bug: empty-state flicker','Detail for Bug: empty-state flicker.','in_review','low',0,'2026-05-27 00:00:00',60,5.50,'{}','2026-03-13 00:00:00','2026-06-27 16:20:00'),(8,'019f09e1-7dc7-7272-b99a-ebfe9e3ff956','019f09e1-7dc7-7272-b99a-ebfa860248ef',NULL,NULL,'Dark-mode token migration','Detail for Dark-mode token migration.','todo','low',0,'2026-06-16 00:00:00',70,7.50,'{}','2026-03-14 00:00:00','2026-06-27 16:20:00'),(9,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','019f09e1-7dc7-7272-b99a-ebf3fd700a5f',NULL,'API v2 design doc','Detail for API v2 design doc.','done','urgent',0,'2026-04-16 00:00:00',80,9.50,'{}','2026-03-15 00:00:00','2026-06-27 16:20:00'),(10,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebfc373fc01b','019f09e1-7dc7-7272-b99a-ebf3fd700a5f',9,'Pagination cursor spec','Detail for Pagination cursor spec.','in_review','high',0,'2026-05-24 00:00:00',90,1.50,'{}','2026-03-16 00:00:00','2026-06-27 16:20:00'),(11,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',9,'OpenAPI 3.1 codegen','Detail for OpenAPI 3.1 codegen.','in_progress','high',0,'2026-06-02 00:00:00',100,3.50,'{}','2026-03-17 00:00:00','2026-06-27 16:20:00'),(12,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebf2ed933f35','019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'Rate limiter (sliding window)','Detail for Rate limiter (sliding window).','todo','high',0,'2026-06-09 00:00:00',110,5.50,'{}','2026-03-18 00:00:00','2026-06-27 16:20:00'),(13,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',NULL,'GraphQL N+1 audit','Detail for GraphQL N+1 audit.','todo','medium',0,'2026-06-25 00:00:00',120,7.50,'{}','2026-03-19 00:00:00','2026-06-27 16:20:00'),(14,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'Bug: 502 on long-polling','Detail for Bug: 502 on long-polling.','in_progress','urgent',0,'2026-05-26 00:00:00',130,9.50,'{}','2026-03-20 00:00:00','2026-06-27 16:20:00'),(15,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebf5bde55420',NULL,NULL,'Webhook signature v2','Detail for Webhook signature v2.','todo','medium',0,'2026-06-20 00:00:00',140,1.50,'{}','2026-03-21 00:00:00','2026-06-27 16:20:00'),(16,'019f09e1-7dc7-7272-b99a-ebfffbda1e2e','019f09e1-7dc7-7272-b99a-ebf62a8f2178','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',NULL,'Drop legacy /v1/sessions','Detail for Drop legacy /v1/sessions.','cancelled','low',0,'2026-05-16 00:00:00',150,3.50,'{}','2026-03-22 00:00:00','2026-06-27 16:20:00'),(17,'019f09e1-7dc7-7272-b99a-ec004c2014e5','019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',NULL,'iOS app scaffolding','Detail for iOS app scaffolding.','done','urgent',0,'2026-04-11 00:00:00',160,5.50,'{}','2026-03-23 00:00:00','2026-06-27 16:20:00'),(18,'019f09e1-7dc7-7272-b99a-ec004c2014e5','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d',NULL,'Android app scaffolding','Detail for Android app scaffolding.','done','urgent',0,'2026-04-11 00:00:00',170,7.50,'{}','2026-03-24 00:00:00','2026-06-27 16:20:00'),(19,'019f09e1-7dc7-7272-b99a-ec004c2014e5','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d',NULL,'Push notification plumbing','Detail for Push notification plumbing.','in_progress','high',0,'2026-06-05 00:00:00',180,9.50,'{}','2026-03-25 00:00:00','2026-06-27 16:20:00'),(20,'019f09e1-7dc7-7272-b99a-ec004c2014e5','019f09e1-7dc7-7272-b99a-ebfa860248ef','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',NULL,'Offline-first cache','Detail for Offline-first cache.','todo','high',0,'2026-06-23 00:00:00',190,1.50,'{}','2026-03-26 00:00:00','2026-06-27 16:20:00'),(21,'019f09e1-7dc7-7272-b99a-ec004c2014e5','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d',NULL,'Crash reporting integration','Detail for Crash reporting integration.','in_review','medium',0,'2026-05-23 00:00:00',200,3.50,'{}','2026-03-27 00:00:00','2026-06-27 16:20:00'),(22,'019f09e1-7dc7-7272-b99a-ec004c2014e5','019f09e1-7dc7-7272-b99a-ebfc373fc01b','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf',NULL,'Bug: scroll jank on iOS 17','Detail for Bug: scroll jank on iOS 17.','todo','medium',0,'2026-06-01 00:00:00',210,5.50,'{}','2026-03-28 00:00:00','2026-06-27 16:20:00'),(23,'019f09e1-7dc7-7272-b99a-ec004c2014e5','019f09e1-7dc7-7272-b99a-ebfda8a94e2c',NULL,NULL,'Deep-link routing','Detail for Deep-link routing.','todo','low',0,'2026-06-30 00:00:00',220,7.50,'{}','2026-03-29 00:00:00','2026-06-27 16:20:00'),(24,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','019f09e1-7dc7-7272-b99a-ebf2ed933f35','019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'GitHub Actions matrix','Detail for GitHub Actions matrix.','done','high',0,'2026-04-06 00:00:00',230,9.50,'{}','2026-03-30 00:00:00','2026-06-27 16:20:00'),(25,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'Docker compose dev env','Detail for Docker compose dev env.','done','high',0,'2026-04-08 00:00:00',240,1.50,'{}','2026-03-31 00:00:00','2026-06-27 16:20:00'),(26,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'Observability stack (loki)','Detail for Observability stack (loki).','in_progress','high',0,'2026-05-31 00:00:00',250,3.50,'{}','2026-04-01 00:00:00','2026-06-27 16:20:00'),(27,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','019f09e1-7dc7-7272-b99a-ebf5bde55420','019f09e1-7dc7-7272-b99a-ebf2ed933f35',NULL,'Secrets rotation runbook','Detail for Secrets rotation runbook.','todo','medium',0,'2026-06-15 00:00:00',260,5.50,'{}','2026-04-02 00:00:00','2026-06-27 16:20:00'),(28,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','019f09e1-7dc7-7272-b99a-ebf62a8f2178','019f09e1-7dc7-7272-b99a-ebf752763e91',NULL,'Postgres → MySQL benchmark','Detail for Postgres → MySQL benchmark.','todo','low',0,'2026-07-10 00:00:00',270,7.50,'{}','2026-04-03 00:00:00','2026-06-27 16:20:00'),(29,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','019f09e1-7dc7-7272-b99a-ebf752763e91','019f09e1-7dc7-7272-b99a-ebf62a8f2178',NULL,'Bug: flaky e2e on retry','Detail for Bug: flaky e2e on retry.','in_progress','high',0,'2026-05-28 00:00:00',280,9.50,'{}','2026-04-04 00:00:00','2026-06-27 16:20:00'),(30,'019f09e1-7dc7-7272-b99a-ec01b3ac2367','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','019f09e1-7dc7-7272-b99a-ebf2ed933f35',NULL,'Tame CI cost','Detail for Tame CI cost.','todo','medium',0,'2026-07-25 00:00:00',290,1.50,'{}','2026-04-05 00:00:00','2026-06-27 16:20:00');
/*!40000 ALTER TABLE `tasks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `time_entries`
--

DROP TABLE IF EXISTS `time_entries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `time_entries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `user_id` char(36) NOT NULL,
  `started_at` datetime NOT NULL,
  `ended_at` datetime DEFAULT NULL COMMENT 'Null while timer is running',
  `minutes_logged` int DEFAULT NULL COMMENT 'Denormalized cache derived from started_at/ended_at',
  `note` text,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_time_entries_task` (`task_id`),
  KEY `idx_time_entries_user` (`user_id`),
  CONSTRAINT `chk_time_entries_ended_after_started` CHECK (((`ended_at` is null) or (`ended_at` >= `started_at`))),
  CONSTRAINT `chk_time_entries_minutes_non_negative` CHECK (((`minutes_logged` is null) or (`minutes_logged` >= 0)))
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `time_entries`
--

LOCK TABLES `time_entries` WRITE;
/*!40000 ALTER TABLE `time_entries` DISABLE KEYS */;
INSERT INTO `time_entries` VALUES (1,1,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-15 09:00:00','2026-05-15 09:45:00',45,'Work session #1','2026-05-15 09:45:00'),(2,2,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-16 10:00:00','2026-05-16 11:00:00',60,'Work session #2','2026-05-16 11:00:00'),(3,3,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-17 11:00:00','2026-05-17 12:15:00',75,'Work session #3','2026-05-17 12:15:00'),(4,4,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-18 12:00:00','2026-05-18 13:30:00',90,'Work session #4','2026-05-18 13:30:00'),(5,5,'019f09e1-7dc7-7272-b99a-ebf5bde55420','2026-05-19 09:00:00','2026-05-19 09:45:00',45,'Work session #5','2026-05-19 09:45:00'),(6,6,'019f09e1-7dc7-7272-b99a-ebf5bde55420','2026-05-20 10:00:00','2026-05-20 11:00:00',60,'Work session #6','2026-05-20 11:00:00'),(7,7,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','2026-05-21 11:00:00','2026-05-21 12:15:00',75,'Work session #7','2026-05-21 12:15:00'),(8,9,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-16 09:00:00','2026-05-16 09:45:00',45,'Work session #8','2026-05-16 09:45:00'),(9,10,'019f09e1-7dc7-7272-b99a-ebf3fd700a5f','2026-05-17 10:00:00','2026-05-17 11:00:00',60,'Work session #9','2026-05-17 11:00:00'),(10,11,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-05-18 11:00:00','2026-05-18 12:15:00',75,'Work session #10','2026-05-18 12:15:00'),(11,12,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-05-19 12:00:00','2026-05-19 13:30:00',90,'Work session #11','2026-05-19 13:30:00'),(12,13,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-05-20 09:00:00','2026-05-20 09:45:00',45,'Work session #12','2026-05-20 09:45:00'),(13,14,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-05-21 10:00:00','2026-05-21 11:00:00',60,'Work session #13','2026-05-21 11:00:00'),(14,16,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-05-16 12:00:00','2026-05-16 13:30:00',90,'Work session #14','2026-05-16 13:30:00'),(15,17,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-05-17 09:00:00','2026-05-17 09:45:00',45,'Work session #15','2026-05-17 09:45:00'),(16,18,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','2026-05-18 10:00:00','2026-05-18 11:00:00',60,'Work session #16','2026-05-18 11:00:00'),(17,19,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','2026-05-19 11:00:00','2026-05-19 12:15:00',75,'Work session #17','2026-05-19 12:15:00'),(18,20,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-05-20 12:00:00','2026-05-20 13:30:00',90,'Work session #18','2026-05-20 13:30:00'),(19,21,'019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','2026-05-21 09:00:00','2026-05-21 09:45:00',45,'Work session #19','2026-05-21 09:45:00'),(20,22,'019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','2026-05-15 10:00:00','2026-05-15 11:00:00',60,'Work session #20','2026-05-15 11:00:00'),(21,24,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-05-17 12:00:00','2026-05-17 13:30:00',90,'Work session #21','2026-05-17 13:30:00'),(22,25,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-05-18 09:00:00','2026-05-18 09:45:00',45,'Work session #22','2026-05-18 09:45:00'),(23,26,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-05-19 10:00:00','2026-05-19 11:00:00',60,'Work session #23','2026-05-19 11:00:00'),(24,27,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-20 11:00:00','2026-05-20 12:15:00',75,'Work session #24','2026-05-20 12:15:00'),(25,28,'019f09e1-7dc7-7272-b99a-ebf752763e91','2026-05-21 12:00:00','2026-05-21 13:30:00',90,'Work session #25','2026-05-21 13:30:00'),(26,29,'019f09e1-7dc7-7272-b99a-ebf62a8f2178','2026-05-15 09:00:00','2026-05-15 09:45:00',45,'Work session #26','2026-05-15 09:45:00'),(27,30,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-16 10:00:00','2026-05-16 11:00:00',60,'Work session #27','2026-05-16 11:00:00'),(28,1,'019f09e1-7dc7-7272-b99a-ebf2ed933f35','2026-05-26 08:00:00',NULL,NULL,'Currently in progress','2026-05-26 08:00:00');
/*!40000 ALTER TABLE `time_entries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` char(36) NOT NULL,
  `email` varchar(255) NOT NULL,
  `full_name` varchar(160) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `avatar_url` varchar(512) DEFAULT NULL,
  `timezone` varchar(64) NOT NULL DEFAULT 'UTC',
  `default_project_id` char(36) DEFAULT NULL COMMENT 'Singular FK twist: column singular, points at projects.id',
  `default_workspace_id` char(36) DEFAULT NULL COMMENT 'Singular FK twist: column singular, points at workspaces.id',
  `last_known_location` point /*!80003 SRID 4326 */ DEFAULT NULL COMMENT 'MySQL 8-exclusive: per-column SRID enforcement',
  `notification_prefs` set('email','push','sms','digest','mentions') NOT NULL DEFAULT 'email,mentions' COMMENT 'MySQL/MariaDB-only SET type — bitfield of opt-in notification channels',
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  `last_login_at` datetime DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES ('019f09e1-7dc7-7272-b99a-ebf2ed933f35','arif@dira.test','Arif Mahmud','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ebfe9e3ff956','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebf3fd700a5f','farzana@dira.test','Farzana Rahman','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ebfffbda1e2e','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','imran@dira.test','Imran Khan','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ec004c2014e5','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebf5bde55420','nadia@dira.test','Nadia Chowdhury','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ec01b3ac2367','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebf62a8f2178','rifat@dira.test','Rifat Sultana','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ebfe9e3ff956','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebf752763e91','sabbir@dira.test','Sabbir Ahmed','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ebfffbda1e2e','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','tania@dira.test','Tania Akter','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ec004c2014e5','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','zubayer@dira.test','Zubayer Hossain','$argon2id$stub',NULL,'Asia/Dhaka','019f09e1-7dc7-7272-b99a-ec01b3ac2367','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebfa860248ef','leo@dira.test','Leo Martin','$argon2id$stub',NULL,'Europe/Berlin','019f09e1-7dc7-7272-b99a-ebfe9e3ff956','00000001-0000-7000-8000-000000000002',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebfbcb51b57a','maya@dira.test','Maya Singh','$argon2id$stub',NULL,'Asia/Kolkata','019f09e1-7dc7-7272-b99a-ebfffbda1e2e','00000001-0000-7000-8000-000000000002',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebfc373fc01b','omar@dira.test','Omar Yilmaz','$argon2id$stub',NULL,'Europe/Istanbul','019f09e1-7dc7-7272-b99a-ec004c2014e5','00000001-0000-7000-8000-000000000002',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00'),('019f09e1-7dc7-7272-b99a-ebfda8a94e2c','priya@dira.test','Priya Patel','$argon2id$stub',NULL,'Asia/Kolkata','019f09e1-7dc7-7272-b99a-ec01b3ac2367','00000001-0000-7000-8000-000000000001',NULL,'email,mentions',1,'2026-05-25 00:00:00','2026-01-26 00:00:00','2026-06-27 16:20:00');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `workspace_members`
--

DROP TABLE IF EXISTS `workspace_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `workspace_members` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `workspace_id` char(36) NOT NULL,
  `user_id` char(36) NOT NULL,
  `role` enum('owner','admin','member','guest') NOT NULL DEFAULT 'member',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_workspace_members` (`workspace_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `workspace_members`
--

LOCK TABLES `workspace_members` WRITE;
/*!40000 ALTER TABLE `workspace_members` DISABLE KEYS */;
INSERT INTO `workspace_members` VALUES (1,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf2ed933f35','owner','2026-02-01 00:00:00'),(2,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf3fd700a5f','admin','2026-02-01 00:00:00'),(3,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf4fa2b4ebf','member','2026-02-01 00:00:00'),(4,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf5bde55420','member','2026-02-01 00:00:00'),(5,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf62a8f2178','member','2026-02-01 00:00:00'),(6,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf752763e91','admin','2026-02-01 00:00:00'),(7,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf86d3ea3ec','member','2026-02-01 00:00:00'),(8,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebf9ce5dd85d','member','2026-02-01 00:00:00'),(9,'00000001-0000-7000-8000-000000000001','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','guest','2026-02-01 00:00:00'),(10,'00000001-0000-7000-8000-000000000002','019f09e1-7dc7-7272-b99a-ebfa860248ef','owner','2026-02-01 00:00:00'),(11,'00000001-0000-7000-8000-000000000002','019f09e1-7dc7-7272-b99a-ebfbcb51b57a','admin','2026-02-01 00:00:00'),(12,'00000001-0000-7000-8000-000000000002','019f09e1-7dc7-7272-b99a-ebfc373fc01b','member','2026-02-01 00:00:00'),(13,'00000001-0000-7000-8000-000000000002','019f09e1-7dc7-7272-b99a-ebfda8a94e2c','member','2026-02-01 00:00:00');
/*!40000 ALTER TABLE `workspace_members` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `workspaces`
--

DROP TABLE IF EXISTS `workspaces`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `workspaces` (
  `id` char(36) NOT NULL,
  `slug` varchar(64) NOT NULL,
  `name` varchar(160) NOT NULL,
  `plan` enum('free','team','business','enterprise') NOT NULL DEFAULT 'free',
  `settings` json NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_workspaces_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `workspaces`
--

LOCK TABLES `workspaces` WRITE;
/*!40000 ALTER TABLE `workspaces` DISABLE KEYS */;
INSERT INTO `workspaces` VALUES ('00000001-0000-7000-8000-000000000001','dira-engineering','Dira Engineering','team','{\"default_view\": \"board\"}','2025-12-01 00:00:00','2026-05-01 00:00:00'),('00000001-0000-7000-8000-000000000002','acme','ACME Corp','business','{\"default_view\": \"list\"}','2026-01-15 00:00:00','2026-05-15 00:00:00');
/*!40000 ALTER TABLE `workspaces` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping events for database 'dbctx_test_dira'
--

--
-- Dumping routines for database 'dbctx_test_dira'
--

--
-- Current Database: `dbctx_test_dira`
--

USE `dbctx_test_dira`;

--
-- Final view structure for view `custom_field_options`
--

/*!50001 DROP VIEW IF EXISTS `custom_field_options`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_unicode_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`root`@`%` SQL SECURITY DEFINER */
/*!50001 VIEW `custom_field_options` AS select `cf`.`id` AS `custom_field_id`,`cf`.`workspace_id` AS `workspace_id`,`cf`.`key` AS `field_key`,`opt`.`position` AS `position`,`opt`.`value` AS `value`,`opt`.`label` AS `label` from (`custom_fields` `cf` join json_table(`cf`.`options`, '$[*]' columns (`position` for ordinality, `value` varchar(255) character set utf8mb4 collate utf8mb4_unicode_ci path '$.value', `label` varchar(255) character set utf8mb4 collate utf8mb4_unicode_ci path '$.label')) `opt`) where (`cf`.`field_type` = 'select') */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-06-27 16:20:12
