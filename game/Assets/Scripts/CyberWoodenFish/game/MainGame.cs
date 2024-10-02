using System.Collections;
using System.Collections.Generic;
using System.Linq;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace CyberWoodenFish.game
{
    public class MainGame : MonoBehaviour
    {

        private int _combo;

        private int _gongde;
        
        [SerializeField] private TMP_Text comboText;
        
        [SerializeField] private TMP_Text gongdeText;

        [SerializeField] private Button chitang;

        [SerializeField] private GameObject subText;

        [SerializeField] private AudioSource sb;
        
        [SerializeField] private AudioSource fish;
        
        private const float MoveSpeed = 2f; // Speed of the text movement
        
        private const float Lifetime = 0.5f; // How long the text stays visible
        
        private const float ComboHideDelay = 1f;

        private bool _boost;
        
        private float _lastComboUpdateTime;
        
        // Start is called once before the first execution of Update after the MonoBehaviour is created
        private void Start()
        { 
            _gongde = PlayerPrefs.GetInt("gongde"); // default 0
            _combo = 0;
            _lastComboUpdateTime = Time.time;

            _boost = PlayerPrefs.GetInt("boost") != 0;
            
            chitang.onClick.AddListener(() =>
            {
                if (_boost)
                {
                    _gongde -= 3;
                    _combo += 3;
                }
                _gongde -= 1;
                _combo += 1;
                _lastComboUpdateTime = Time.time;
                var text = Instantiate(subText);
                text.SetActive(true);
                StartCoroutine(MoveAndDestroyText(text));
            });
        }

        // Update is called once per frame
        private void Update()
        {
            // Update gongde text
            if (_gongde.ToString() != GetNumber(gongdeText))
            {
                PlayerPrefs.SetInt("gongde", _gongde);
                SetNumber(gongdeText, "当前功德", _gongde);
            }

            // Hide combo text if no update in the last second
            if (Time.time - _lastComboUpdateTime >= ComboHideDelay)
            {
                comboText.gameObject.SetActive(false);
            }
            else
            {
                // Show combo text if it's not visible and _combo is not zero
                if (_combo <= 0) return;
                comboText.gameObject.SetActive(true);
                if (_combo.ToString() != GetNumber(comboText))
                {
                    SetNumber(comboText, "Combo", _combo);
                }
            }
        }
        
        private IEnumerator MoveAndDestroyText(GameObject text)
        {

            if (Random.Range(0, 30) >= 5)
            {
                sb.Play();
            }
            else
            {
                fish.Play();
            }
            
            var rectTransform = text.gameObject.GetComponent<RectTransform>();
            var elapsedTime = 0f;

            while (elapsedTime < Lifetime)
            {
                var yOffset = MoveSpeed * Time.deltaTime;
                rectTransform.anchoredPosition += new Vector2(0, yOffset);
                elapsedTime += Time.deltaTime;
                yield return null;
            }

            Destroy(text);
        }

        private static string GetNumber(TMP_Text text)
        {
            return text.text.Split(": ")[1];
        }

        private static void SetNumber(TMP_Text text, string type,int number)
        {
            text.text = $"{type}: {number}";
        }
    }
}
