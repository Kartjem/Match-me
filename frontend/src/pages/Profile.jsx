import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import "./Profile.css";

const countryList = ["USA", "Canada", "UK", "Mexico", "Germany", "Estonia"];

const cityOptionsMap = {
  USA: ["New York", "Los Angeles", "Chicago", "Houston", "Dallas"],
  Canada: ["Toronto", "Vancouver", "Montreal", "Calgary"],
  UK: ["London", "Manchester", "Liverpool", "Birmingham"],
  Mexico: ["Mexico City", "Guadalajara", "Monterrey"],
  Germany: ["Berlin", "Hamburg", "Munich", "Frankfurt"],
  Estonia: ["Tallinn", "Tartu", "Narva", "Pärnu"],
};

const hobbyOptions = [
  "Reading",
  "Gaming",
  "Cooking",
  "Art",
  "Sports",
  "Music",
  "Travel",
  "Photography",
];

const interestOptions = [
  "Movies",
  "Music",
  "Sports",
  "Coding",
  "Nature",
  "Pets",
  "Art",
  "Theatre",
];

function Profile() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    email: "",
    fname: "",
    surname: "",
    gender: "male",
    birthdate: "",
    about: "",
    hobbies: [],
    interests: [],
    country: "",
    city: "",
    looking_for_gender: "any",
    looking_for_min_age: 18,
    looking_for_max_age: 99,
    profile_picture_url: "",
    preferred_hobbies: [],
    preferred_interests: [],
  });
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);

  useEffect(() => {
    async function fetchProfile() {
      try {
        const res = await fetch("http://localhost:8080/me", { credentials: "include" });
        if (res.status === 403 || res.status === 401) {
          navigate("/login");
          return;
        }
        if (!res.ok) {
          console.error("Failed to fetch profile");
          return;
        }
        const data = await res.json();
        setFormData({
          email: data.email || "",
          fname: data.fname || "",
          surname: data.surname || "",
          gender: data.gender || "male",
          birthdate: data.birthdate || "",
          about: data.about || "",
          hobbies: data.hobbies || [],
          interests: data.interests || [],
          country: data.country || "",
          city: data.city || "",
          looking_for_gender: data.looking_for_gender || "any",
          looking_for_min_age: data.looking_for_min_age || 18,
          looking_for_max_age: data.looking_for_max_age || 99,
          profile_picture_url: data.profile_picture_url || "",
          preferred_hobbies: data.preferred_hobbies || [],
          preferred_interests: data.preferred_interests || [],
        });
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    fetchProfile();
  }, [navigate]);

  useEffect(() => {
    const validCities = cityOptionsMap[formData.country] || [];
    if (!validCities.includes(formData.city)) {
      setFormData((prev) => ({ ...prev, city: "" }));
    }
  }, [formData.country, formData.city]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const toggleHobby = (hobby) => {
    setFormData((prev) => {
      const alreadySelected = prev.hobbies.includes(hobby);
      return {
        ...prev,
        hobbies: alreadySelected ? prev.hobbies.filter((h) => h !== hobby) : [...prev.hobbies, hobby],
      };
    });
  };

  const toggleInterest = (interest) => {
    setFormData((prev) => {
      const alreadySelected = prev.interests.includes(interest);
      return {
        ...prev,
        interests: alreadySelected ? prev.interests.filter((i) => i !== interest) : [...prev.interests, interest],
      };
    });
  };

  const togglePreferredHobby = (hobby) => {
    setFormData((prev) => {
      const alreadySelected = prev.preferred_hobbies.includes(hobby);
      return {
        ...prev,
        preferred_hobbies: alreadySelected
          ? prev.preferred_hobbies.filter((h) => h !== hobby)
          : [...prev.preferred_hobbies, hobby],
      };
    });
  };

  const togglePreferredInterest = (interest) => {
    setFormData((prev) => {
      const alreadySelected = prev.preferred_interests.includes(interest);
      return {
        ...prev,
        preferred_interests: alreadySelected
          ? prev.preferred_interests.filter((i) => i !== interest)
          : [...prev.preferred_interests, interest],
      };
    });
  };

  const handleSave = async () => {
    try {
      const payload = {
        fname: formData.fname || null,
        surname: formData.surname || null,
        gender: formData.gender || null,
        birthdate: formData.birthdate || null,
        about: formData.about || null,
        hobbies: formData.hobbies.length > 0 ? formData.hobbies : null,
        interests: formData.interests.length > 0 ? formData.interests : null,
        country: formData.country || null,
        city: formData.city || null,
        looking_for_gender: formData.looking_for_gender || "any",
        looking_for_min_age: formData.looking_for_min_age ? parseInt(formData.looking_for_min_age, 10) : 18,
        looking_for_max_age: formData.looking_for_max_age ? parseInt(formData.looking_for_max_age, 10) : 99,
        profile_picture_url: formData.profile_picture_url || null,
        preferred_hobbies: formData.preferred_hobbies.length > 0 ? formData.preferred_hobbies : null,
        preferred_interests: formData.preferred_interests.length > 0 ? formData.preferred_interests : null,
      };
      const res = await fetch("http://localhost:8080/update-profile", {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        alert("Profile updated successfully!");
        setEditing(false);
      } else {
        const errText = await res.text();
        alert("Error updating profile: " + errText);
      }
    } catch (err) {
      console.error(err);
      alert("Error saving profile. Check console.");
    }
  };

  if (loading)
    return <div className="profile-container">Loading profile...</div>;

  if (!editing) {
    return (
      <div className="profile-container">
        <h2>My Profile</h2>
        <div className="profile-card">
          <div className="profile-picture">
            {formData.profile_picture_url ? (
              <img src={formData.profile_picture_url} alt="Profile" />
            ) : (
              <div className="placeholder">👤</div>
            )}
          </div>
          <div className="profile-info">
            <p><strong>Email:</strong> {formData.email}</p>
            <p><strong>First Name:</strong> {formData.fname}</p>
            <p><strong>Last Name:</strong> {formData.surname}</p>
            <p><strong>Gender:</strong> {formData.gender}</p>
            <p><strong>Birthdate:</strong> {formData.birthdate}</p>
            <p><strong>Country:</strong> {formData.country}</p>
            <p><strong>City:</strong> {formData.city}</p>
            <p><strong>About:</strong> {formData.about}</p>
            <p>
              <strong>Hobbies:</strong>{" "}
              {Array.isArray(formData.hobbies) ? formData.hobbies.join(", ") : formData.hobbies}
            </p>
            <p>
              <strong>Interests:</strong>{" "}
              {Array.isArray(formData.interests) ? formData.interests.join(", ") : formData.interests}
            </p>
            <p><strong>Looking for:</strong> {formData.looking_for_gender}</p>
            <p>
              <strong>Age Range:</strong> {formData.looking_for_min_age} - {formData.looking_for_max_age}
            </p>
            <p>
              <strong>Preferred Hobbies:</strong>{" "}
              {Array.isArray(formData.preferred_hobbies) ? formData.preferred_hobbies.join(", ") : formData.preferred_hobbies}
            </p>
            <p>
              <strong>Preferred Interests:</strong>{" "}
              {Array.isArray(formData.preferred_interests) ? formData.preferred_interests.join(", ") : formData.preferred_interests}
            </p>
          </div>
        </div>
        <button className="edit-btn" onClick={() => setEditing(true)}>Edit Profile</button>
      </div>
    );
  }

  return (
    <div className="profile-edit-container">
      <h2>Edit Profile</h2>
      <div className="edit-form">
        {/* Basic profile fields */}
        <div className="form-section">
          <label>First Name</label>
          <input type="text" name="fname" value={formData.fname} onChange={handleChange} />
        </div>
        <div className="form-section">
          <label>Last Name</label>
          <input type="text" name="surname" value={formData.surname} onChange={handleChange} />
        </div>
        <div className="form-section">
          <label>Gender</label>
          <select name="gender" value={formData.gender} onChange={handleChange}>
            <option value="">Select Gender</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
            <option value="other">Other</option>
          </select>
        </div>
        <div className="form-section">
          <label>Birthdate</label>
          <input type="date" name="birthdate" value={formData.birthdate} onChange={handleChange} />
        </div>
        <div className="form-section">
          <label>Country</label>
          <select name="country" value={formData.country} onChange={handleChange}>
            <option value="">Select Country</option>
            {countryList.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        </div>
        <div className="form-section">
          <label>City</label>
          <select name="city" value={formData.city} onChange={handleChange}>
            <option value="">Select City</option>
            {(cityOptionsMap[formData.country] || []).map((city) => (
              <option key={city} value={city}>{city}</option>
            ))}
          </select>
        </div>
        <div className="form-section">
          <label>About Me</label>
          <textarea name="about" value={formData.about} onChange={handleChange} />
        </div>
        <div className="form-section checkbox-group">
          <label>Hobbies</label>
          {hobbyOptions.map((hobby) => (
            <div key={hobby} className="checkbox-item">
              <input type="checkbox" checked={formData.hobbies.includes(hobby)} onChange={() => toggleHobby(hobby)} />
              <span>{hobby}</span>
            </div>
          ))}
        </div>
        <div className="form-section checkbox-group">
          <label>Interests</label>
          {interestOptions.map((interest) => (
            <div key={interest} className="checkbox-item">
              <input type="checkbox" checked={formData.interests.includes(interest)} onChange={() => toggleInterest(interest)} />
              <span>{interest}</span>
            </div>
          ))}
        </div>
        <div className="form-section checkbox-group">
          <label>Preferred Hobbies (Partner should have these)</label>
          {hobbyOptions.map((hobby) => (
            <div key={hobby} className="checkbox-item">
              <input type="checkbox" checked={formData.preferred_hobbies.includes(hobby)} onChange={() => togglePreferredHobby(hobby)} />
              <span>{hobby}</span>
            </div>
          ))}
        </div>
        <div className="form-section checkbox-group">
          <label>Preferred Interests (Partner should have these)</label>
          {interestOptions.map((interest) => (
            <div key={interest} className="checkbox-item">
              <input type="checkbox" checked={formData.preferred_interests.includes(interest)} onChange={() => togglePreferredInterest(interest)} />
              <span>{interest}</span>
            </div>
          ))}
        </div>
        <div className="form-section">
          <label>Looking for Gender</label>
          <select name="looking_for_gender" value={formData.looking_for_gender} onChange={handleChange}>
            <option value="any">Any</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
            <option value="other">Other</option>
          </select>
        </div>
        <div className="form-section">
          <label>Looking for Min Age</label>
          <input type="number" name="looking_for_min_age" value={formData.looking_for_min_age} onChange={handleChange} />
        </div>
        <div className="form-section">
          <label>Looking for Max Age</label>
          <input type="number" name="looking_for_max_age" value={formData.looking_for_max_age} onChange={handleChange} />
        </div>
        <div className="form-section">
          <label>Profile Picture URL</label>
          <input type="text" name="profile_picture_url" value={formData.profile_picture_url} onChange={handleChange} />
        </div>
        <div className="form-buttons">
          <button className="save-btn" onClick={handleSave}>Save</button>
          <button className="cancel-btn" onClick={() => setEditing(false)}>Cancel</button>
        </div>
      </div>
    </div>
  );
}

export default Profile;