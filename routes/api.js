// routes/api.js
const express = require('express');
const jwt = require('jsonwebtoken');
const User = require('../models/User');
const { authenticateJWT } = require('../middleware/authenticateJWT');

const router = express.Router();

/**
 * @swagger
 * tags:
 *   name: JSON Files
 *   description: API for managing JSON files
 */

/**
 * @swagger
 * /files:
 *   get:
 *     summary: Get all JSON files for authenticated user
 *     tags: [JSON Files]
 *     responses:
 *       200:
 *         description: Successfully retrieved JSON files
 *         content:
 *           application/json:
 *             schema:
 *               type: array
 *               items:
 *                 type: object
 */
router.get('/files', authenticateJWT, async (req, res) => {
    try {
        const user = await User.findById(req.user.id);
        res.json(user.jsonFiles);
    } catch (err) {
        res.status(500).send('Server error');
    }
});

/**
 * @swagger
 * /files:
 *   post:
 *     summary: Create a new JSON file for authenticated user
 *     tags: [JSON Files]
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               name:
 *                 type: string
 *               content:
 *                 type: object
 *     responses:
 *       201:
 *         description: Successfully created JSON file
 */
router.post('/files', authenticateJWT, async (req, res) => {
    const { name, content } = req.body;
    try {
        const user = await User.findById(req.user.id);
        user.jsonFiles.push({ name, content });
        await user.save();
        res.status(201).json(user.jsonFiles);
    } catch (err) {
        res.status(500).send('Server error');
    }
});

/**
 * @swagger
 * /files/{id}:
 *   delete:
 *     summary: Delete a JSON file for authenticated user
 *     tags: [JSON Files]
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *     responses:
 *       200:
 *         description: Successfully deleted JSON file
 */
router.delete('/files/:id', authenticateJWT, async (req, res) => {
    try {
        const user = await User.findById(req.user.id);
        user.jsonFiles = user.jsonFiles.filter(file => file._id.toString() !== req.params.id);
        await user.save();
        res.status(200).json(user.jsonFiles);
    } catch (err) {
        res.status(500).send('Server error');
    }
});

module.exports = router;
