-- SQLite

-- --get post author
-- SELECT UserID FROM Posts
-- WHERE Posts.PostID= "4d9bd266-6866-4be1-8ea1-4c1f85efe088";

-- INSERT INTO Notifications
-- VALUES (?,?,?,?)

-- SELECT UserID FROM Posts WHERE Posts.ParentID = "27a1be9f-51f1-4c81-ad76-087254518459" LIMIT 1;

-- DELETE FROM Notifications WHERE Notifications.UserID = "410856a0-4ffb-4a09-9403-58a186dfe242";
-- DELETE FROM Posts WHERE Posts.UserID = "a568b82f-f987-4c1f-8027-73eb0ed9380b";

-- SELECT UserID FROM Posts WHERE Posts.ParentID= "27a1be9f-51f1-4c81-ad76-087254518459" LIMIT 1;

-- SELECT * FROM Notifications WHERE UserID = '7bcc27c3-4a95-4c1a-b72f-1db0c93978f7';

-- SELECT Notifications.UserID, Users.UserName, Notifications.PostID,Notifications.Type, Posts.ParentID, Posts.PostTitle
-- FROM Notifications 
-- LEFT JOIN Users ON Notifications.ReactorID = Users.UserID 
-- LEFT JOIN Posts ON Notifications.PostID = Posts.PostID
-- WHERE Notifications.UserID = "7bcc27c3-4a95-4c1a-b72f-1db0c93978f7";