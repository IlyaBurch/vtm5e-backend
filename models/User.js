// models/User.js
const mongoose = require('mongoose');

const UserSchema = new mongoose.Schema({
    googleId: {
        type: String,
        required: true,
    },
    displayName: {
        type: String,
        required: true,
    },
    jsonFiles: [{
        name: String,
        content: Object,
    }],
});

module.exports = mongoose.model('User', UserSchema);
